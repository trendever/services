package models

import (
	"core/api"
	"core/chat"
	"core/db"
	"errors"
	"fmt"
	proto_chat "proto/chat"
	proto_core "proto/core"
	"utils/log"
	"utils/rpc"
)

func init() {
	RegisterTemplate("chat", "product_buy")
	RegisterTemplate("chat", "product_info")
}

var templatesMap = map[proto_core.LeadAction]string{
	proto_core.LeadAction_BUY:  "product_buy",
	proto_core.LeadAction_INFO: "product_info",
}

//SendProductToChat sends the product to the lead chat
func SendProductToChat(lead *Lead, product *Product, action proto_core.LeadAction, source string) error {
	log.Debug("SendProductToChat(%v, %v, %v)", lead.ID, product.ID, action)

	// determine whether product is special
	var specials []uint
	res := db.New().
		Select("DISTINCT product_id").
		Table("chat_templates").
		Where("product_id IS NOT NULL").
		Pluck("product_id", &specials)
	if res.Error != nil {
		return errors.New("failed to load special products list")
	}
	isSpecial := false
	for _, s := range specials {
		if s == product.ID {
			isSpecial = true
			break
		}
	}

	// test for repetitive actions
	var count uint
	res = db.New().
		Select("COUNT(1)").
		Table("products_leads lead").
		Joins("JOIN products_leads_items l_item ON l_item.lead_id = lead.id").
		Joins("JOIN products_product_item p_item ON l_item.product_item_id = p_item.id").
		Where("lead.customer_id = ?", lead.Customer.ID).
		Where("lead.id <> ?", lead.ID)
	if isSpecial {
		res = res.Where("p_item.product_id = ?", product.ID)
	} else {
		if len(specials) > 0 {
			res = res.Where("p_item.product_id NOT IN (?)", specials)
		}
	}
	err := res.Row().Scan(&count)
	log.Debug("count = %v; %v", count, err)
	if err != nil {
		return fmt.Errorf("failed to determine whether user is new: %v", err)
	}
	// load templates
	var templates []struct {
		ChatTemplateMessage
		TemplateName string
		ProductID    uint
	}
	res = db.New().
		Select("msg.text, msg.position, tmpl.template_name, tmpl.product_id").
		Table("chat_template_messages msg").
		Joins("JOIN chat_template_cases c ON c.id = msg.case_id").
		Joins("JOIN chat_templates tmpl ON tmpl.id = c.template_id").
		Where("tmpl.group = ?", templatesMap[action]).
		Where("tmpl.product_id = ? OR tmpl.is_default", product.ID).
		Where("c.for_suppliers_with_notices = ?", product.Shop.NotifySupplier).
		Where("c.for_new_users = ? ", count == 0).
		Where("c.source = ?", source).
		Order("tmpl.product_id IS NULL, msg.position").
		Scan(&templates)
	if res.Error != nil {
		return fmt.Errorf("failed to load templates: %v", res.Error)
	}
	if len(templates) == 0 {
		return fmt.Errorf(
			"suitable tamplates not found for productID = %v with action %v",
			product.ID,
			proto_core.LeadAction_name[int32(action)],
		)
	}

	err = joinChat(lead.ConversationID, &SystemUser, proto_chat.MemberRole_SYSTEM)
	if err != nil {
		return fmt.Errorf("failed to join chat: %v", err)
	}

	specific := templates[0].ProductID

	for _, tmpl := range templates {
		// we are at end of specific templates(if any)
		if tmpl.ProductID != specific {
			break
		}

		content, err := tmpl.Execute(map[string]interface{}{
			"lead":    lead,
			"product": product,
		})

		if err != nil {
			log.Error(fmt.Errorf(
				"failed to parse template with id %v for product %v in lead %v: %v",
				tmpl.ID, product.ID, lead.ID, err,
			))
			continue
		}

		str, ok := content.(string)
		if !ok {
			log.Error(fmt.Errorf("template '%v' returned unexpected type", tmpl.TemplateName))
			continue
		}
		if str == "" {
			log.Warn("template '%v' returned empty string", tmpl.TemplateName)
			continue
		}

		err = chat.SendChatMessage(uint64(SystemUser.ID), lead.ConversationID, content.(string), "text/html")
		if err != nil {
			return fmt.Errorf("failed to send message to chat: %v", err)
		}
	}
	return nil
}

//SendStatusMessage sends status message
func SendStatusMessage(conversationID uint64, statusType, value string) {
	content := &chat.StatusContent{
		Type:  statusType,
		Value: value,
	}
	m := &proto_chat.Message{
		ConversationId: conversationID,
		Parts: []*proto_chat.MessagePart{
			{
				Content:  content.JSON(),
				MimeType: "json/status",
			},
		},
	}
	api.Publish("chat.status", m)
}

func joinChat(conversationID uint64, user *User, role proto_chat.MemberRole) error {
	context, cancel := rpc.DefaultContext()
	defer cancel()

	members := []*proto_chat.Member{{
		UserId: uint64(user.ID), Name: user.GetName(), Role: role,
	}}
	resp, err := api.ChatServiceClient.JoinChat(
		context,
		&proto_chat.JoinChatRequest{
			ConversationId: conversationID,
			Members:        members,
		},
	)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("JoinChat method returned error %v: %v", resp.Error.Code, resp.Error.Message)
	}
	return nil
}
