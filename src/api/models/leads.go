package models

import (
	"proto/chat"
	"proto/core"
)

type Lead struct {
	core.LeadInfo
	Chat           *chat.Chat `json:"chat"`
	ConversationId uint64     `json:"conversation_id,omitempty"` //dirty hack, it removes this field from json
}

type Leads []*Lead

func (l *Leads) Fill(leads []*core.LeadInfo, chats []*chat.Chat) {
	chatMap := map[uint64]*chat.Chat{}
	for _, chat := range chats {
		chatMap[chat.Id] = chat
	}

	for _, lead := range leads {
		chat, _ := chatMap[lead.ConversationId]
		*l = append(*l, &Lead{
			LeadInfo: *lead,
			Chat:     chat,
		})
	}
}
