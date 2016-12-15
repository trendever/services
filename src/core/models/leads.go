package models

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"github.com/qor/validations"
	"proto/core"
	"time"
	"utils/db"
)

// Possible lead sources
var LeadSources = []string{
	"website",
	"wantit",
}

// Lead model
type Lead struct {
	gorm.Model

	// One of LeadSources
	Source string

	CustomerID uint
	Customer   User

	ShopID uint
	Shop   Shop

	ProductItems []ProductItem `gorm:"many2many:products_leads_items"`
	Products     []*Product    `sql:"-"`
	// user comment from instagram
	Comment       string `gorm:"text"`
	InstagramPk   string `gorm:"index"`
	InstagramLink string // link to reposted instagram product

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
	return l.State == "NEW" || l.State == "EMPTY"
}

//Encode returns LeadInfo
func (l *Lead) Encode() *core.LeadInfo {
	state, _ := core.LeadStatus_value[l.State]
	lead := &core.LeadInfo{
		Id:             uint64(l.ID),
		Source:         l.Source,
		CustomerId:     uint64(l.CustomerID),
		InstagramPk:    l.InstagramPk,
		InstagramLink:  l.InstagramLink,
		Status:         core.LeadStatus(state),
		ConversationId: l.ConversationID,
		UserRole:       l.UserRole,
		UpdatedAt:      l.ChatUpdatedAt.UnixNano(),
		UpdatedAtAgo:   int64(time.Since(l.ChatUpdatedAt).Seconds()),
		CancelReason:   uint64(l.CancelReasonID.Int64),
		StatusComment:  l.StatusComment,
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

		InstagramPk:   lead.InstagramPk,
		InstagramLink: lead.InstagramLink,
		Comment:       lead.Comment,
	}

}

func (lead *Lead) TriggerEvent(event string) error {
	err := LeadState.Trigger(event, lead, db.New())
	if err != nil {
		return fmt.Errorf("failed to trigger event: %v", err)
	}
	err = db.New().Model(&lead).UpdateColumn("state", lead.State).Error
	if err != nil {
		return fmt.Errorf("failed to update lead state: %v", err)
	}
	return nil
}
