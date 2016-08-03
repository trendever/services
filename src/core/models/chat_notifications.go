package models

import (
	"core/api"
	"core/chat"
	"core/db"
	"fmt"
	proto_chat "proto/chat"
	proto_core "proto/core"
	"utils/log"
	"utils/rpc"
)

var templatesMap = map[proto_core.LeadAction]string{
	proto_core.LeadAction_BUY:  "product_buy",
	proto_core.LeadAction_INFO: "product_info",
}

//SendProductToChat sends the product to the lead chat
func SendProductToChat(lead *Lead, product *Product, action proto_core.LeadAction) error {
	log.Debug("SendProductToChat(%v, %v, %v)", lead.ID, product.ID, action)
	var templates []ChatTemplate
	res := db.New().
		Where(`"group" = ?`, templatesMap[action]).
		Where("product_id = ? OR is_default", product.ID).
		Where("for_suppliers_with_notices = ?", product.Shop.NotifySupplier).
		Order(`product_id desc, "order"`).
		Find(&templates)
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

	err := joinChat(lead.ConversationID, &SystemUser, proto_chat.MemberRole_SYSTEM)
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
