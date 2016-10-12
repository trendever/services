package views

import (
	"core/api"
	"core/models"
	"core/telegram"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/chat"
	"proto/core"
	"utils/db"
	"utils/log"
)

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		core.RegisterLeadServiceServer(s, leadServer{
			notifier: models.GetNotifier(),
		})
	})
}

type leadServer struct {
	notifier *models.Notifier
}

func (s leadServer) CreateLead(ctx context.Context, protoLead *core.Lead) (*core.CreateLeadResult, error) {

	var err error
	if protoLead.ProductId == 0 {
		return nil, errors.New("ProductID is required")
	}

	var lead *models.Lead

	var product *models.Product
	if product, err = models.GetProductByID(uint64(protoLead.ProductId), "Items", "InstagramImages"); err != nil {
		log.Error(err)
		return nil, err
	}
	prod := product
	existsLead, err := models.FindActiveLead(uint64(prod.ShopID), uint64(protoLead.CustomerId))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	//Create new lead if lead not exists, or use exists
	if existsLead == nil {
		lead, err = models.CreateLead(protoLead, prod.ShopID)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else {
		lead = existsLead
	}

	if models.LeadEventPossible(core.LeadStatusEvent_CREATE.String(), lead.State) {
		//Event CREATE performs chat creation
		if err := models.LeadState.Trigger(core.LeadStatusEvent_CREATE.String(), lead, db.New()); err == nil {
			//this errors not critical, we can change status from EMPTY to NEW later
			err = db.New().Model(lead).UpdateColumn("state", lead.State).Error
			if err != nil {
				log.Error(err)
			}
		} else {
			//that's also not critical
			log.Error(err)
		}
	}

	if protoLead.Action == core.LeadAction_BUY {
		if count, err := models.AppendLeadItems(lead, prod.Items); err != nil {
			log.Error(err)
			return nil, err
		} else if count == 0 {
			// This mean the product already in the lead, or product without items,
			// anyway we don't want to do anything more with this lead
			leadInfo, err := models.GetUserLead(&lead.Customer, uint64(lead.ID))
			if err != nil {
				log.Error(err)
				return nil, err
			}

			return &core.CreateLeadResult{
				Id:   int64(lead.ID),
				Lead: leadInfo.Encode(),
			}, nil
		}
	}

	// If chat is down, conversation is not created (yet)
	// Later CREATE lead event (see below) can be triggered to fix it
	// So, everything is partly fine now
	if lead.ConversationID != 0 {
		go func() {
			log.Error(models.SendProductToChat(lead, prod, protoLead.Action, protoLead.Source, false))
		}()
	} else {
		log.Error(errors.New("lead.ConversationID == 0"))
	}

	go telegram.NotifyLeadCreated(lead, prod, protoLead.InstagramLink, protoLead.Action)
	// @CHECK may be it's wrong place to do it
	if existsLead != nil {
		// send this message only on new lead
		go NotifyAboutLeadEvent(lead, "CREATE")
	}

	leadInfo, err := models.GetUserLead(&lead.Customer, uint64(lead.ID))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &core.CreateLeadResult{
		Id:   int64(lead.ID),
		Lead: leadInfo.Encode(),
	}, nil
}

// ReadLead checks lead existence by id
func (s leadServer) ReadLead(ctx context.Context, req *core.ReadLeadRequest) (*core.ReadLeadResult, error) {

	searchLead := models.Lead{
		Model: gorm.Model{
			ID: uint(req.GetId()),
		},
		InstagramPk:    req.GetInstagramPk(),
		ConversationID: req.GetConversationId(),
	}

	query := db.New().Where(&searchLead).Find(&searchLead)
	if query.Error != nil && !query.RecordNotFound() {
		log.Error(query.Error)
		return nil, query.Error
	}

	return &core.ReadLeadResult{
		Id: int64(searchLead.ID),
	}, nil
}

//GetUserLeads returns user leads filtred by user's roles
func (s leadServer) GetUserLeads(ctx context.Context, req *core.UserLeadsRequest) (reply *core.UserLeadsReply, err error) {
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}

	user, err := models.GetUserByID(uint(req.UserId))

	if err != nil {
		log.Error(err)
		return nil, err
	}

	reply = &core.UserLeadsReply{
		Leads: []*core.LeadInfo{},
	}
	leads, err := models.GetUserLeads(user, req.Role, req.LeadId, req.Limit, req.FromUpdatedAt, req.Direction)
	if err != nil {
		log.Error(err)
		return
	}
	reply.Leads = leads.Encode()
	return
}

//SetLeadStatus sets new lead status
func (s leadServer) SetLeadStatus(ctx context.Context, req *core.SetLeadStatusRequest) (reply *core.SetLeadStatusReply, err error) {

	user, err := models.GetUserByID(uint(req.UserId))

	if err != nil {
		log.Error(err)
		return nil, err
	}

	var lead *models.Lead
	lead, err = models.GetUserLead(user, req.LeadId)
	if err != nil {
		log.Error(err)
		return
	}

	if !models.CanUserChangeLeadState(lead.UserRole.String(), req.Event.String()) {
		log.Debug("User %v with role %s can't set lead %v state to %v", req.UserId, lead.UserRole.String(), req.LeadId, req.Event.String())
		return nil, errors.New("Forbidden")
	}

	upd := map[string]interface{}{
		"cancel_reason_id": sql.NullInt64{},
	}
	reason := models.LeadCancelReason{ID: req.CancelReason}
	reasonIsValid := false
	if req.Event == core.LeadStatusEvent_CANCEL {
		err := db.New().First(&reason).Error
		if err != nil {
			log.Errorf("failed to load cancel reason %v: %v", reason.ID, err)
		} else {
			upd["cancel_reason_id"] = reason.ID
			reasonIsValid = true
		}

	}

	err = models.LeadState.Trigger(req.Event.String(), lead, db.New())
	upd["state"] = lead.State
	if err != nil {
		err = fmt.Errorf("failed to trigger lead event %v: %v", req.Event, err)
		log.Error(err)
		return nil, err
	}

	if req.StatusComment != "" {
		upd["status_comment"] = req.StatusComment
	}

	if err := db.New().Model(lead).UpdateColumns(upd).Error; err != nil {
		log.Error(err)
		return nil, err
	}

	if reasonIsValid {
		chatMsg, err := reason.GenChatMessage(lead, user)
		if err != nil {
			log.Errorf(
				"failed to generate chat message for cancel reason %v: %v",
				reason.ID, err,
			)
		}
		if chatMsg != "" {
			go func() {
				log.Error(models.SendChatMessages(
					lead.ConversationID,
					&chat.Message{
						UserId: uint64(models.SystemUser.ID),
						Parts: []*chat.MessagePart{
							{Content: string(chatMsg), MimeType: "text/plain"},
						},
					},
				))
			}()
		}
	}

	// notify stuff
	go models.SendStatusMessage(lead.ConversationID, "lead.state.changed", lead.State)
	go NotifyAboutLeadEvent(lead, req.Event.String())

	return &core.SetLeadStatusReply{Lead: lead.Encode()}, nil
}

//CallSupplier calls supplier to the chat
func (s leadServer) CallSupplier(ctx context.Context, req *core.CallSupplierRequest) (reply *core.CallSupplierReply, err error) {
	lead := &models.Lead{}
	if lead, err = models.GetFullLeadByID(req.LeadId); err != nil {
		return
	}

	supplier := lead.Shop.Supplier

	if supplier.Phone == "" {
		log.Warn("Supplier doesn't have phone number. Lead: %v, Supllier: %v", lead.ID, lead.Shop.SupplierID)
		return nil, errors.New("Supplier doesn't have phone number")
	}

	go func() {
		log.Error(s.notifier.CallSupplierToChat(&supplier, lead))
	}()
	go models.SendStatusMessage(lead.ConversationID, "suplier.called", "")

	return &core.CallSupplierReply{}, nil

}

//GetLead returns full lead info
func (s leadServer) GetLead(_ context.Context, req *core.GetLeadRequest) (*core.GetLeadReply, error) {
	if req.UserId == 0 {
		return nil, errors.New("User id is required")
	}
	user, err := models.GetUserByID(uint(req.UserId))

	if err != nil {
		log.Error(err)
		return nil, err
	}
	searchLead := models.Lead{
		Model: gorm.Model{
			ID: uint(req.GetId()),
		},
		ConversationID: req.GetConversationId(),
	}

	query := db.New().Where(&searchLead).Find(&searchLead)
	if query.Error != nil && !query.RecordNotFound() {
		log.Error(query.Error)
		return nil, query.Error
	}

	if query.RecordNotFound() {
		return &core.GetLeadReply{}, nil
	}

	lead, err := models.GetUserLead(user, uint64(searchLead.ID))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &core.GetLeadReply{
		Lead: lead.Encode(),
	}, nil
}

//CallCustomer calls customer to the chat
func (s leadServer) CallCustomer(_ context.Context, req *core.CallCustomerRequest) (reply *core.CallCustomerReply, err error) {
	lead := &models.Lead{}
	if lead, err = models.GetFullLeadByID(req.LeadId); err != nil {
		log.Error(err)
		return
	}

	customer := lead.Customer

	if customer.Phone == "" {
		return nil, errors.New("Customer doesn't have phone number")
	}

	go func() {
		log.Error(s.notifier.CallCustomerToChat(&customer, lead))
	}()
	go models.SendStatusMessage(lead.ConversationID, "customer.called", "")
	return &core.CallCustomerReply{}, nil
}

func (s leadServer) GetCancelReasons(_ context.Context, in *core.GetCancelReasonsRequest) (*core.GetCancelReasonsReply, error) {
	var reasons []models.LeadCancelReason
	err := db.New().Find(&reasons).Error
	if err != nil {
		return nil, err
	}
	ret := &core.GetCancelReasonsReply{Reasons: make([]*core.CancelReason, 0, len(reasons))}
	for _, reason := range reasons {
		ret.Reasons = append(ret.Reasons, &core.CancelReason{
			Id:   reason.ID,
			Name: reason.Name,
		})
	}
	return ret, nil
}
