package models

import (
	"core/api"
	"encoding/json"
	"errors"
	"fmt"
	"proto/chat"
	"proto/core"
	"utils/db"
	"utils/log"
	"utils/rpc"
)

func init() {
	RegisterTemplate("chat", "product_buy")
	RegisterTemplate("chat", "product_info")
	RegisterTemplate("chat", "product_chat_init")
}

var templatesMap = map[core.LeadAction]string{
	core.LeadAction_BUY:  "product_buy",
	core.LeadAction_INFO: "product_info",
}

//SendProductToChat sends the product to the lead chat
func SendProductToChat(lead *Lead, product *Product, action core.LeadAction, source string, chat_init bool) error {
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
	if err != nil {
		return fmt.Errorf("failed to determine whether user is new: %v", err)
	}

	err = SendChatTemplates(templatesMap[action], lead, product, count == 0, source)
	if err == nil && chat_init {
		err = SendChatTemplates("product_chat_init", lead, product, count == 0, source)
	}
	return err
}

func SendChatTemplates(group string, lead *Lead, product *Product, isNewUser bool, source string) error {
	// load templates
	var templates []struct {
		ChatTemplateMessage
		TemplateName string
		ProductID    uint
	}
	res := db.New().
		Select("msg.text, msg.data, msg.position, tmpl.template_name, tmpl.product_id").
		Table("chat_template_messages msg").
		Joins("JOIN chat_templates tmpl ON tmpl.id = msg.template_id").
		Where("tmpl.group = ?", group).
		Where("tmpl.product_id = ? OR tmpl.is_default", product.ID).
		Order("tmpl.product_id IS NULL, msg.position").
		Scan(&templates)
	if res.Error != nil {
		return fmt.Errorf("failed to load templates: %v", res.Error)
	}
	if len(templates) == 0 {
		log.Errorf(
			"suitable tamplates not found for productID = %v with group %v",
			product.ID,
			group,
		)
		return nil
	}

	err := joinChat(lead.ConversationID, chat.MemberRole_SYSTEM, &SystemUser)
	if err != nil {
		return fmt.Errorf("failed to join chat: %v", err)
	}

	specific := templates[0].ProductID

	messages := []*chat.Message{}
	for _, tmpl := range templates {
		// we are at end of specific templates(if any)
		if tmpl.ProductID != specific {
			break
		}

		content, err := tmpl.Execute(map[string]interface{}{
			"lead":    lead,
			"product": product,
			"source":  source,
			"newUser": isNewUser,
		})

		if err != nil {
			log.Errorf(
				"failed to parse template with id %v for product %v in lead %v: %v",
				tmpl.ID, product.ID, lead.ID, err,
			)
			continue
		}

		parts, ok := content.([]*chat.MessagePart)
		if !ok {
			log.Errorf("template '%v' returned unexpected type", tmpl.TemplateName)
			continue
		}
		if len(parts) == 0 {
			log.Warn("template '%v' returned empty result", tmpl.TemplateName)
			continue
		}
		log.Debug("%v parts: %+v", len(parts), parts)
		messages = append(messages, &chat.Message{
			UserId: uint64(SystemUser.ID),
			Parts:  parts,
		})
	}
	err = SendChatMessages(lead.ConversationID, messages...)
	if err != nil {
		return fmt.Errorf("failed to send messages to chat: %v", err)
	}
	return nil
}

// SendChatMessages sends messages to chat
func SendChatMessages(conversationID uint64, messages ...*chat.Message) error {
	if len(messages) == 0 {
		return nil
	}
	context, cancel := rpc.DefaultContext()
	defer cancel()
	_, err := api.ChatServiceClient.SendNewMessage(context, &chat.SendMessageRequest{
		ConversationId: conversationID,
		Messages:       messages,
	})
	return err
}

//SendStatusMessage sends status message
func SendStatusMessage(conversationID uint64, statusType, value string) {
	err := joinChat(conversationID, chat.MemberRole_SYSTEM, &SystemUser)
	if err != nil {
		log.Errorf("failed to join chat: %v", err)
		return
	}
	content, _ := json.Marshal(struct {
		Type  string `json:"type"`
		Value string `json:"value,omitempty"`
	}{
		Type:  statusType,
		Value: value,
	})
	err = SendChatMessages(
		conversationID,
		&chat.Message{
			UserId: uint64(SystemUser.ID),
			Parts: []*chat.MessagePart{
				{Content: string(content), MimeType: "json/status"},
			},
		},
	)
	if err != nil {
		log.Errorf("failed to send message: %v", err)
	}
}

func joinChat(conversationID uint64, role chat.MemberRole, users ...*User) error {
	if conversationID == 0 || len(users) == 0 {
		return nil
	}
	context, cancel := rpc.DefaultContext()
	defer cancel()

	var members []*chat.Member
	for _, user := range users {
		members = append(members, &chat.Member{
			UserId: uint64(user.ID),
			Name:   user.GetName(),
			Role:   role,
		})
	}
	resp, err := api.ChatServiceClient.JoinChat(
		context,
		&chat.JoinChatRequest{
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
