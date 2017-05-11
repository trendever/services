package models

import (
	"core/api"
	"core/conf"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"proto/core"
	"utils/log"
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
			leadEventCancel,
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
	RegisterTemplate("other", "chat_caption")

	LeadState.Initial(leadStateNew)

	// init state machine
	for _, state := range leadStates {
		LeadState.State(state)
	}

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
						log.Errorf("failed to send notify for supplier: %v", err)
					}
				}

				for _, seller := range shop.Sellers {
					if err := n.NotifySellerAboutLead(seller, lead); err != nil {
						log.Errorf("failed to send notify for supplier: %v", err)
					}
				}
			}()
			go api.NotifyByTelegram(
				api.TelegramChannelNewLead,
				fmt.Sprintf(
					"Lead %v from %v to %v in now active.\n%v\n%v",
					lead.ID, lead.Customer.DisplayName(), lead.Shop.Stringify(),
					fmt.Sprintf("%v/chat/%v", conf.GetSettings().SiteURL, lead.ID),
					fmt.Sprintf("%v/qor/orders/%v", conf.GetSettings().SiteURL, lead.ID),
				),
			)
			return nil
		},
	)

	LeadState.State(leadStateSubmited).Enter(
		func(model interface{}, tx *gorm.DB) error {
			lead, ok := model.(*Lead)
			if !ok {
				return errors.New("Unsuported type for lead")
			}
			if lead.Source == "website" {
				return nil
			}
			return SetChatSync(lead.ConversationID, "")
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
