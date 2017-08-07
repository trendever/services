package views

import (
	"core/api"
	"core/conf"
	"core/models"
	"errors"
	"fmt"
	"proto/core"
	"strings"
	"utils/db"
	"utils/log"

	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
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

// @REFACTOR split/simplify this func somehow?
func (s leadServer) CreateLead(ctx context.Context, protoLead *core.Lead) (*core.CreateLeadResult, error) {

	var err error
	if protoLead.ProductId == 0 {
		return nil, errors.New("ProductID is required")
	}

	var lead *models.Lead

	var product *models.Product
	if product, err = models.GetProductByID(uint64(protoLead.ProductId), "Items", "InstagramImages", "Shop"); err != nil {
		log.Error(err)
		return nil, err
	}

	//Skipping leads with comments that not match cfg
	if protoLead.Source == "comment" {
		lead_skipped := models.Lead{}.Decode(protoLead)
		comment_prepared := strings.ToLower(protoLead.Comment)
		is_comment_matched := false

		vocabulary := conf.GetSettings().Comments.Allowed

		//Match comment if it contains key phrases
		for _, phrase := range strings.Split(vocabulary, ",") {
			phrase_stemmed, _ := models.PrepareText(phrase, "russian")
			if strings.Contains(comment_prepared, phrase_stemmed) || strings.Contains(comment_prepared, "?") {
				is_comment_matched = true
				break
			}
		}

		//Also match if shop name supplied in comment
		if strings.Contains(comment_prepared, fmt.Sprintf("%s", product.Shop.InstagramUsername)) {
			is_comment_matched = true
		}

		if !is_comment_matched {
			//notify about skiped lead
			go models.NotifyLeadCreated(lead_skipped, product, protoLead.InstagramLink, core.LeadAction_SKIP)
			//prevent next steps
			return &core.CreateLeadResult{}, nil
		}
	}

	existsLead, err := models.FindActiveLead(uint64(product.ShopID), uint64(protoLead.CustomerId), uint64(protoLead.ProductId))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	//Create new lead if lead not exists, or use exists
	if existsLead == nil {
		lead, err = models.CreateLead(protoLead, product.ShopID)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else {
		lead = existsLead

		lead.Comment = protoLead.Comment
		lead.InstagramMediaId = protoLead.InstagramMediaId
		lead.Source = protoLead.Source

		err := db.New().Model(lead).Updates(map[string]string{
			"Comment":          lead.Comment,
			"InstagramMediaId": lead.InstagramMediaId,
			"Source":           lead.Source,
		}).Error

		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	// comment leads should be auto-advances
	if lead.Source == "comment" && models.LeadEventPossible(core.LeadStatusEvent_PROGRESS.String(), lead.State) {
		log.Error(lead.TriggerEvent(core.LeadStatusEvent_PROGRESS.String(), "", 0, &lead.Customer))
	}

	if protoLead.Action == core.LeadAction_BUY {
		if count, err := models.AppendLeadItems(lead, product.Items); err != nil {
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

	go func() {
		if protoLead.Source == "direct" {
			models.SetChatSync(lead.ConversationID, protoLead.DirectThread)
		}
		if err := models.SendProductToChat(lead, product, protoLead.Action, protoLead.Source, protoLead.Comment, existsLead == nil); err != nil {
			log.Error(fmt.Errorf("Warning! Could not send product to chat (%v)", err))
		}
	}()

	go models.NotifyLeadCreated(lead, product, protoLead.InstagramLink, protoLead.Action)

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
		log.Errorf("User %v with role %s can't set lead %v state to %v", req.UserId, lead.UserRole.String(), req.LeadId, req.Event.String())
		return nil, errors.New("Forbidden")
	}

	err = lead.TriggerEvent(req.Event.String(), req.StatusComment, req.CancelReason, user)
	if err != nil {
		err = fmt.Errorf("failed to trigger lead event %v: %v", req.Event, err)
		log.Error(err)
		return nil, err
	}

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

func (s leadServer) GetUserRole(_ context.Context, in *core.GetUserRoleRequest) (ret *core.GetUserRoleReply, _ error) {
	ret = &core.GetUserRoleReply{}

	if (in.LeadId == 0 && in.ConversationId == 0) || (in.UserId == 0 && in.InstagramUserId == 0) {
		ret.Error = "empty conditions"
		return
	}

	user, found, err := models.FindUserMatchAny(in.UserId, in.InstagramUserId, "", "", "", "")
	if err != nil {
		ret.Error = fmt.Sprintf("failed to load user: %v", err)
		return
	}
	if !found {
		ret.Error = "user not found"
		return
	}

	lead := models.Lead{
		Model: gorm.Model{
			ID: uint(in.LeadId),
		},
		ConversationID: in.ConversationId,
	}

	res := db.New().Preload("Shop").Preload("Shop.Sellers").Where(&lead).First(&lead)
	if res.RecordNotFound() {
		ret.Error = "lead not found"
		return
	}
	if res.Error != nil {
		ret.Error = fmt.Sprintf("failed to load lead: %v", res.Error)
		return
	}
	switch {
	case lead.CustomerID == user.ID:
		ret.Role = core.LeadUserRole_CUSTOMER

	case lead.Shop.SupplierID == user.ID:
		ret.Role = core.LeadUserRole_SUPPLIER

	case lead.Shop.HasSeller(user.ID):
		ret.Role = core.LeadUserRole_SELLER

	case user.SuperSeller:
		ret.Role = core.LeadUserRole_SUPER_SELLER

	default:
		ret.Role = core.LeadUserRole_UNKNOWN
	}
	return
}
