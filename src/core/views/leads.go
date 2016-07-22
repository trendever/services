package views

import (
	"core/api"
	"core/db"
	"core/models"
	"core/notifier"
	"core/telegram"
	"errors"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
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
	notifier notifier.Notifier
}

// Decodes protobuf model to core model
//  product items are not fully loaded
func decodeLead(l *core.Lead) *models.Lead {
	lead := &models.Lead{
		Source: l.Source,

		CustomerID: uint(l.CustomerId),

		InstagramPk: l.InstagramPk,
		Comment:     l.Comment,
	}

	return lead
}

func (s leadServer) CreateLead(ctx context.Context, protoLead *core.Lead) (*core.CreateLeadResult, error) {

	var err error
	if protoLead.ProductId == 0 {
		log.Warn("Can't cread lead without product id")
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

	go telegram.NotifyLeadCreated(lead, prod, protoLead.InstagramLink)

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

	if lead.ConversationID != 0 {
		//send only if lead already existed. Because if this is a new lead, we already send product to chat
		//in CREATE event handler
		if existsLead != nil {
			go log.Error(models.SendProductToChat(lead, prod))
		}
		//notify sellers
		go func() {
			shop := &models.Shop{}
			err := db.New().Model(&models.Shop{}).Preload("Sellers").Preload("Supplier").Find(shop, prod.ShopID).Error
			if err != nil {
				log.Error(err)
				return
			}

			if shop.NotifySupplier {
				s.notifySellerAboutLead(lead, shop.Supplier)
			}

			for _, seller := range shop.Sellers {
				s.notifySellerAboutLead(lead, seller)
			}
		}()
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

// ReadLead checks lead existance by id
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

	reply = &core.SetLeadStatusReply{}
	var lead *models.Lead
	lead, err = models.GetUserLead(user, req.LeadId)
	if err != nil {
		log.Error(err)
		return
	}
	if models.CanUserChangeLeadState(lead.UserRole.String(), req.Event.String()) {
		err = models.LeadState.Trigger(req.Event.String(), lead, db.New())
		if err == nil {
			if err := db.New().Model(lead).UpdateColumn("state", lead.State).Error; err != nil {
				log.Error(err)
			} else {
				go models.SendStatusMessage(lead.ConversationID, "lead.state.changed", lead.State)
			}
		} else {
			log.Error(err)
		}
	} else {
		log.Debug("User role %s, event %s, user %v, lead %v", lead.UserRole.String(), req.Event.String(), req.UserId, req.LeadId)
		err = errors.New("Forbidden")
	}
	reply.Lead = lead.Encode()
	return
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

	token, err := api.GetNewAPIToken(supplier.ID)
	if err != nil {
		log.Warn("Can't get token for supplier: %v %v: %v", supplier.ID, supplier.Phone, err)
		return nil, errors.New("Can't get token for supplier")
	}
	url := api.GetChatURL(lead.ID, token)
	result, err := api.GetShortURL(url)
	if err != nil {
		log.Error(err)
	} else {
		url = result.URL
		//sends only short url
		go func() {
			err := notifier.CallSupplierToChat(supplier, url, lead, s.notifier.NotifyBySms)
			if err != nil {
				log.Error(err)
			}
		}()
	}

	if supplier.Email != "" {
		go func() {
			err := notifier.CallSupplierToChat(supplier, url, lead, s.notifier.NotifyByEmail)
			if err != nil {
				log.Error(err)
			}
		}()
	}

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

	token, err := api.GetNewAPIToken(customer.ID)
	if err != nil {
		log.Error(err)
		return nil, errors.New("Can't get token for customer")
	}
	url := api.GetChatURL(lead.ID, token)
	result, err := api.GetShortURL(url)
	if err != nil {
		log.Error(err)
	} else {
		url = result.URL
		//sends only short url
		go func() {
			err := notifier.CallCustomerToChat(customer, url, lead, s.notifier.NotifyBySms)
			if err != nil {
				log.Error(err)
			}
		}()
	}

	if customer.Email != "" {
		go func() {
			err := notifier.CallCustomerToChat(customer, url, lead, s.notifier.NotifyByEmail)
			if err != nil {
				log.Error(err)
			}
		}()
	}

	go models.SendStatusMessage(lead.ConversationID, "customer.called", "")
	return &core.CallCustomerReply{}, nil
}

func (s leadServer) notifySellerAboutLead(lead *models.Lead, seller models.User) {
	if seller.Phone == "" {
		return
	}

	url, err := api.GetChatURLWithToken(lead.ID, seller.ID)
	if err != nil {
		log.Error(err)
		return
	}

	short, err := api.GetShortURL(url)
	if err == nil {
		err := notifier.NotifySellerAboutLead(seller, short.URL, lead, s.notifier.NotifyBySms)
		if err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}

	if seller.Email != "" {
		err := notifier.NotifySellerAboutLead(seller, url, lead, s.notifier.NotifyByEmail)
		if err != nil {
			log.Error(err)
		}
	}
}
