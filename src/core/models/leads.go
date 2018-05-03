package models

import (
	"common/db"
	"common/log"
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"github.com/qor/validations"
	"proto/chat"
	"proto/core"
	"time"
	"utils/nats"
)

// Possible lead sources
var LeadSources = []string{
	"website",
	"wantit",
	"comment",
}

// Lead model
type Lead struct {
	gorm.Model

	// One of LeadSources
	Source string

	CustomerID uint `gorm:"index"`
	Customer   User `gorm:"save_associations:false"`

	ShopID uint `gorm:"index"`
	Shop   Shop `gorm:"save_associations:false"`

	ProductItems []ProductItem `gorm:"many2many:products_leads_items"`
	Products     []*Product    `sql:"-"`
	// user comment from instagram
	Comment          string `gorm:"text"`
	InstagramPk      string `gorm:"index"`
	InstagramLink    string // link to reposted instagram product
	InstagramMediaId string // Id of post where lead originated

	ConversationID uint64 `gorm:"index"`

	transition.Transition

	UserRole core.LeadUserRole `sql:"-"`

	//IsNotified is a bool flag, true - means customer was notified about this lead
	IsNotified bool `gorm:"index"`
	//ChatUpdatedAt is a field for sorting leads
	ChatUpdatedAt time.Time `gorm:"index"`

	CancelReasonID sql.NullInt64
	CancelReason   LeadCancelReason `gorm:"ForeignKey:CancelReasonID"`
	// comment about lead status
	StatusComment string `gorm:"text"`
}

// Validate lead
func (l Lead) Validate(db *gorm.DB) {
	stateOk := false
	for _, realState := range leadStates {
		if l.State == realState {
			stateOk = true
			break
		}
	}

	if l.State == "" {
		stateOk = true
	}

	if !stateOk {
		db.AddError(validations.NewError(Lead{}, "State", "Incorrect state"))
	}
}

// TableName for this model
func (l Lead) TableName() string {
	return "products_leads"
}

func (l Lead) IsNew() bool {
	return l.State == leadStateNew || l.State == leadStateEmpty
}

//Encode returns LeadInfo
func (l *Lead) Encode() *core.LeadInfo {
	state, _ := core.LeadStatus_value[l.State]
	lead := &core.LeadInfo{
		Id:               uint64(l.ID),
		Source:           l.Source,
		CustomerId:       uint64(l.CustomerID),
		InstagramPk:      l.InstagramPk,
		InstagramLink:    l.InstagramLink,
		InstagramMediaId: l.InstagramMediaId,
		Status:           core.LeadStatus(state),
		ConversationId:   l.ConversationID,
		UserRole:         l.UserRole,
		UpdatedAt:        l.ChatUpdatedAt.UnixNano(),
		UpdatedAtAgo:     int64(time.Since(l.ChatUpdatedAt).Seconds()),
		CancelReason:     uint64(l.CancelReasonID.Int64),
		StatusComment:    l.StatusComment,
	}

	//if l.ProductItems != nil {
	//	lead.Items = []*core.ProductItem{}
	//	for _, item := range l.ProductItems {
	//		lead.Items = append(lead.Items, item.ToLeadInfoItem())
	//	}
	//}

	if l.Products != nil {
		lead.Products = Products(l.Products).Encode()
	}

	if &l.Shop != nil {
		lead.Shop = l.Shop.Encode()
	}

	if &l.Customer != nil {
		lead.Customer = l.Customer.PublicEncode()
	}

	return lead
}

//LeadCollection collection of leads
type LeadCollection []*Lead

//Encode returns encoded collection
func (ls LeadCollection) Encode() []*core.LeadInfo {
	leads := []*core.LeadInfo{}
	for _, item := range ls {
		lead := item.Encode()
		leads = append(leads, lead)

	}
	return leads
}

//Decode decodes core.Lead to Lead
func (l Lead) Decode(lead *core.Lead) *Lead {
	return &Lead{
		Source: lead.Source,

		CustomerID: uint(lead.CustomerId),

		InstagramPk:      lead.InstagramPk,
		InstagramLink:    lead.InstagramLink,
		InstagramMediaId: lead.InstagramMediaId,
		Comment:          lead.Comment,
	}

}

// Shop with sellers must be loaded before this call
func (lead Lead) RoleOf(user *User) core.LeadUserRole {
	switch {
	case lead.CustomerID == user.ID:
		return core.LeadUserRole_CUSTOMER
	case lead.Shop.SupplierID == user.ID:
		return core.LeadUserRole_SUPPLIER
	case lead.Shop.HasSeller(user.ID):
		return core.LeadUserRole_SELLER
	case user.SuperSeller:
		return core.LeadUserRole_SUPER_SELLER
	default:
		return core.LeadUserRole_UNKNOWN
	}
}

func (lead *Lead) TriggerEvent(eventName, statusComment string, cancelReason uint64, mover *User) error {
	event, ok := leadEvents[eventName]
	if !ok {
		return fmt.Errorf("unknown event %v", event)
	}
	if event.To == lead.State {
		return nil
	}
	err := LeadState.Trigger(eventName, lead, db.New())
	if err != nil {
		return fmt.Errorf("failed to trigger event: %v", err)
	}

	upd := map[string]interface{}{
		"state":            lead.State,
		"status_comment":   statusComment,
		"cancel_reason_id": sql.NullInt64{},
	}

	reason := LeadCancelReason{ID: cancelReason}
	reasonIsValid := false
	if eventName == leadEventCancel {
		err := db.New().First(&reason).Error
		if err != nil {
			log.Errorf("failed to load cancel reason %v: %v", reason.ID, err)
		} else {
			upd["cancel_reason_id"] = reason.ID
			reasonIsValid = true
		}

	}

	if err := db.New().Model(lead).UpdateColumns(upd).Error; err != nil {
		return err
	}

	if mover != nil {
		if reasonIsValid {
			chatMsg, err := reason.GenChatMessage(lead, mover)
			if err != nil {
				log.Errorf(
					"failed to generate chat message for cancel reason %v: %v",
					reason.ID, err,
				)
			}
			if chatMsg != nil {
				go func() {
					log.Error(SendChatMessages(
						lead.ConversationID,
						chatMsg,
					))
				}()
			}
		}
	}

	// notify stuff
	go SendStatusMessage(lead.ConversationID, "lead.state.changed", lead.State)
	go NotifyAboutLeadEvent(lead, eventName)

	return nil
}

// NotifyAboutLeadEvent notifies about lead event via NATS, changes related conversation status
func NotifyAboutLeadEvent(lead *Lead, event string) {

	log.Debug("Notifying about lead %v event", lead.ID)

	users, err := GetUsersForLead(lead)
	if err != nil {
		log.Errorf("failed to get related users for lead %v: %v", lead.ID, err)
	}

	err = nats.StanPublish("core.lead.event", &core.LeadEventMessage{
		LeadId: uint64(lead.ID),
		Users:  users,
		Event:  event,
	})
	if err != nil {
		log.Errorf("failed to publush core.lead.event: %v", err)
	}

	chatStatus := "new"
	switch lead.State {
	case leadStateNew, leadStateEmpty:

	case leadStateCancelled:
		chatStatus = "cancelled"
	default:
		chatStatus = "active"
	}
	err = nats.StanPublish("chat.conversation.set_status", &chat.SetStatusMessage{
		ConversationId: lead.ConversationID,
		Status:         chatStatus,
	})
	if err != nil {
		log.Errorf("failed to publush chat.conversation.set_status: %v", err)
	}
}
