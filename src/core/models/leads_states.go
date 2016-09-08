package models

import (
	"core/api"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"proto/chat"
	"proto/core"
	"utils/db"
	"utils/log"
	"utils/rpc"
)

// Possible lead states
var (
	leadStateEmpty      = core.LeadStatus_EMPTY.String()
	leadStateNew        = core.LeadStatus_NEW.String()
	leadStateInProgress = core.LeadStatus_IN_PROGRESS.String()
	leadStateSubmited   = core.LeadStatus_SUBMITTED.String()
	leadStateOnDelivery = core.LeadStatus_ON_DELIVERY.String()
	leadStateCompleted  = core.LeadStatus_COMPLETED.String()
	leadStateCancelled  = core.LeadStatus_CANCELLED.String()
)

// Possible lead events
var (
	leadEventCreate   = core.LeadStatusEvent_CREATE.String()
	leadEventProgress = core.LeadStatusEvent_PROGRESS.String()
	leadEventSubmit   = core.LeadStatusEvent_SUBMIT.String()
	leadEventDelivery = core.LeadStatusEvent_DELIVERY.String()
	leadEventComplete = core.LeadStatusEvent_COMPLETE.String()
	leadEventCancel   = core.LeadStatusEvent_CANCEL.String()
)

//Lead roles
var (
	leadRoleCustomer    = core.LeadUserRole_CUSTOMER.String()
	leadRoleSupplier    = core.LeadUserRole_SUPPLIER.String()
	leadRoleSeller      = core.LeadUserRole_SELLER.String()
	leadRoleSuperSeller = core.LeadUserRole_SUPER_SELLER.String()
)

var (
	//LeadEventPermission declares who can trigger events, and what events
	leadEventPermission = map[string][]string{
		leadRoleCustomer: {
			leadEventProgress,
			leadEventCreate,
		},
		leadRoleSeller: {
			leadEventCreate,
			leadEventProgress,
			leadEventSubmit,
			leadEventDelivery,
			leadEventComplete,
			leadEventCancel,
		},
		leadRoleSupplier: {
			leadEventCreate,
			leadEventProgress,
			leadEventSubmit,
			leadEventDelivery,
			leadEventComplete,
			leadEventCancel,
		},
		leadRoleSuperSeller: {
			leadEventCreate,
			leadEventProgress,
			leadEventSubmit,
			leadEventDelivery,
			leadEventComplete,
			leadEventCancel,
		},
	}
	// LeadStates contains all possible lead stated as a slice
	leadStates = []string{
		leadStateEmpty,
		leadStateNew,
		leadStateInProgress,
		leadStateSubmited,
		leadStateOnDelivery,
		leadStateCompleted,
		leadStateCancelled,
	}

	// LeadEvents contains all possible lead events as a slice
	leadEvents = map[string]LeadEvent{
		// event -- toState	--	fromStates
		leadEventCreate: {
			Name: leadEventCreate,
			To:   leadStateNew,
			From: []string{
				leadStateEmpty,
				leadStateNew,
				leadStateInProgress,
				leadStateSubmited,
				leadStateOnDelivery,
				leadStateCompleted,
				leadStateCancelled,
			},
		},
		leadEventProgress: {
			Name: leadEventProgress,
			To:   leadStateInProgress,
			From: []string{
				leadStateEmpty,
				leadStateNew,
				leadStateInProgress,
				leadStateSubmited,
				leadStateOnDelivery,
				leadStateCompleted,
				leadStateCancelled,
			},
		},
		leadEventSubmit: {
			Name: leadEventSubmit,
			To:   leadStateSubmited,
			From: []string{
				leadStateEmpty,
				leadStateNew,
				leadStateInProgress,
				leadStateSubmited,
				leadStateOnDelivery,
				leadStateCompleted,
				leadStateCancelled,
			},
		},
		leadEventDelivery: {
			Name: leadEventDelivery,
			To:   leadStateOnDelivery,
			From: []string{
				leadStateEmpty,
				leadStateNew,
				leadStateInProgress,
				leadStateSubmited,
				leadStateOnDelivery,
				leadStateCompleted,
				leadStateCancelled,
			},
		},
		leadEventComplete: {
			Name: leadEventComplete,
			To:   leadStateCompleted,
			From: []string{
				leadStateEmpty,
				leadStateNew,
				leadStateInProgress,
				leadStateSubmited,
				leadStateOnDelivery,
				leadStateCompleted,
				leadStateCancelled,
			},
		},
		leadEventCancel: {
			Name: leadEventCancel,
			To:   leadStateCancelled,
			From: []string{
				leadStateEmpty,
				leadStateNew,
				leadStateInProgress,
				leadStateSubmited,
				leadStateOnDelivery,
				leadStateCompleted,
				leadStateCancelled,
			},
		},
	}

	// LeadState is a transition controller
	LeadState = transition.New(&Lead{})
)

// LeadEvent -- transition event container
type LeadEvent struct {
	Name string
	To   string
	From []string
}

func init() {
	LeadState.Initial(leadStateNew)

	// init state machine
	for _, state := range leadStates {
		LeadState.State(state)
	}

	// @CHECK @REFACTOR Some ugly err-ignoring spaghetti happens here. WTF?
	// @TODO Use transcation not to set new state if some errors hapenned during state changing?
	LeadState.State(leadStateNew).Enter(func(model interface{}, tx *gorm.DB) error {
		lead, ok := model.(*Lead)
		if !ok {
			return errors.New("Unsuported type for lead")
		}
		//already has related chat
		if lead.ConversationID != 0 {
			return nil
		}
		context, cancel := rpc.DefaultContext()
		defer cancel()
		var members []*chat.Member

		if customer, err := GetUserByID(lead.CustomerID); err == nil {
			members = append(members, &chat.Member{
				UserId: uint64(customer.ID),
				Role:   chat.MemberRole_CUSTOMER,
				Name:   customer.GetName(),
			})
		}

		if sellers, err := GetSellersByShopID(lead.ShopID); err == nil {
			for _, seller := range sellers {
				members = append(members, &chat.Member{
					UserId: uint64(seller.ID),
					Role:   chat.MemberRole_SELLER,
					Name:   seller.GetName(),
				})
			}
		}

		resp, err := api.ChatServiceClient.CreateChat(context, &chat.NewChatRequest{
			Chat: &chat.Chat{
				Members: members,
			},
		})
		if err != nil {
			return err
		}
		lead.ConversationID = resp.Chat.Id
		err = db.New().Model(lead).UpdateColumn("conversation_id", resp.Chat.Id).Error
		if err != nil {
			log.Error(err)
		}
		return nil
	})

	LeadState.State(leadStateInProgress).Enter(
		func(model interface{}, tx *gorm.DB) error {
			lead, ok := model.(*Lead)
			if !ok {
				return errors.New("Unsuported type for lead")
			}
			go func() {
				shop := &lead.Shop
				n := GetNotifier()
				// @CHECK We cann't handle errors here for real...
				if shop.NotifySupplier {
					if err := n.NotifySellerAboutLead(&shop.Supplier, lead); err != nil {
						log.Error(fmt.Errorf(
							"failed to send notify for supplier: %v", err,
						))
					}
				}

				for _, seller := range shop.Sellers {
					if err := n.NotifySellerAboutLead(seller, lead); err != nil {
						log.Error(fmt.Errorf(
							"failed to send notify for supplier: %v", err,
						))
					}
				}
			}()
			return nil
		},
	)

	for _, event := range leadEvents {
		LeadState.Event(event.Name).To(event.To).From(event.From...)
	}
}

// LeadEventPossible returns true if triggering event eventName from specified state is possible
func LeadEventPossible(eventName, state string) bool {
	event, eventOk := leadEvents[eventName]
	if !eventOk {
		return false
	}

	for _, fromState := range event.From {
		if fromState == state {
			return true
		}
	}

	return false
}

//CanUserChangeLeadState checks can user change state of lead to this state or not
func CanUserChangeLeadState(role, state string) bool {
	states, ok := leadEventPermission[role]
	if !ok {
		return false
	}
	for _, s := range states {
		if state == s {
			return true
		}
	}
	return false
}

func hasLeadRole(role core.LeadUserRole, roles []core.LeadUserRole) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

//GetLeadEvents returns possible lead events
func GetLeadEvents() map[string]LeadEvent {
	return leadEvents
}

//GetLeadStates returns possible lead states
func GetLeadStates() []string {
	return leadStates
}

func makeUintMap(arr []uint64) map[uint64]int {
	m := make(map[uint64]int)
	for i, id := range arr {
		m[id] = i
	}
	return m
}
