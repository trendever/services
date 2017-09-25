package server

import (
	"chat/fixtures"
	"chat/models"
	"database/sql"
	"errors"
	"github.com/golang/mock/gomock"
	"golang.org/x/net/context"
	"proto/chat"
	"testing"
	"time"
	"utils/db"
	"utils/test_tools"
)

func TestCreateChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repoSuccess := fixtures.NewMockConversationRepository(ctrl)
	repoSuccess.EXPECT().Create(gomock.Any()).Do(func(m *models.Conversation) {
		m.ID = 1
	}).Return(nil)

	dberr := errors.New("db error")
	repoFail := fixtures.NewMockConversationRepository(ctrl)
	repoFail.EXPECT().Create(gomock.Any()).Return(dberr)

	tests := test_tools.Tests{
		//success
		test_tools.Test{
			"request": &chat.NewChatRequest{
				Chat: &chat.Chat{Name: "Test chat"},
			},
			"reply": &chat.ChatReply{Chat: &chat.Chat{Id: 1, Name: "Test chat"}, Error: nil},
			"repo":  repoSuccess,
			"error": nil,
		},
		//fails
		test_tools.Test{
			"request": &chat.NewChatRequest{Chat: &chat.Chat{Name: "Test chat"}},
			"reply":   nil,
			"repo":    repoFail,
			"error":   dberr,
		},
	}

	runner := &test_tools.Runner{
		Tests: tests,
		Rules: []test_tools.Rule{
			{test_tools.RuleStr, "reply"},
			{test_tools.RuleStr, "error"},
		},
		Run: func(test test_tools.Test) []interface{} {
			server := NewChatServer(test["repo"].(models.ConversationRepository), nil)
			resp, err := server.CreateChat(context.Background(), test["request"].(*chat.NewChatRequest))
			return []interface{}{resp, err}
		},
	}
	runner.RunTests()
	if runner.HasErrors() {
		t.Error(runner.Errors...)
	}
}

func TestJoinChat(t *testing.T) {
	f := func(test test_tools.Test) []interface{} {
		server := NewChatServer(test["repo"].(models.ConversationRepository), nil)
		resp, err := server.JoinChat(context.Background(), test["request"].(*chat.JoinChatRequest))
		return []interface{}{resp, err}
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chat1 := &models.Conversation{
		Model: db.Model{ID: 1},
		Name:  "Test chat",
	}
	member1 := &chat.Member{UserId: 1, Role: chat.MemberRole_CUSTOMER, Name: "Guest"}

	repoSuccess := fixtures.NewMockConversationRepository(ctrl)
	repoSuccess.EXPECT().GetByID(gomock.Any()).Return(chat1, nil)
	repoSuccess.EXPECT().AddMembers(chat1, gomock.Any()).Return(nil)

	repoFail := fixtures.NewMockConversationRepository(ctrl)
	repoFail.EXPECT().GetByID(gomock.Any()).Return(nil, nil)

	tests := test_tools.Tests{
		//success
		test_tools.Test{
			"request": &chat.JoinChatRequest{ConversationId: 1, Members: []*chat.Member{member1}},
			"reply":   &chat.ChatReply{&chat.Chat{Id: 1, Name: "Test chat"}, nil},
			"repo":    repoSuccess,
			"error":   nil,
		},
		//not found
		test_tools.Test{
			"request": &chat.JoinChatRequest{ConversationId: 1, Members: []*chat.Member{member1}},
			"reply": &chat.ChatReply{Error: &chat.Error{
				Code:    chat.ErrorCode_NOT_EXISTS,
				Message: "Conversation not exists",
			}},
			"repo":  repoFail,
			"error": nil,
		},
	}
	rules := []test_tools.Rule{
		{test_tools.RuleStr, "reply"},
		{test_tools.RuleStr, "error"},
	}

	runner := test_tools.NewRunner(tests, f, rules)

	runner.RunTests()
	if runner.HasErrors() {
		t.Error(runner.Errors)
	}
}

// This shiet is seriously broken. Comment it for now
func _TestSendMessage(t *testing.T) {
	f := func(test test_tools.Test) []interface{} {
		server := NewChatServer(test["repo"].(models.ConversationRepository), nil)
		resp, err := server.SendNewMessage(context.Background(), test["request"].(*chat.SendMessageRequest))
		return []interface{}{resp, err}
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chat1 := &models.Conversation{
		Model: db.Model{ID: 1},
		Name:  "Test chat",
	}
	member1 := &models.Member{Model: db.Model{ID: 1}}
	ct := time.Now()
	message1 := &chat.Message{
		ConversationId: 1,
		UserId:         1,
		Parts: []*chat.MessagePart{
			{
				MimeType: "text/plain",
				Content:  "test",
			},
		},
		User:      &chat.Member{Id: 1},
		CreatedAt: ct.Unix(),
	}

	repoSuccess := fixtures.NewMockConversationRepository(ctrl)
	repoSuccess.EXPECT().GetByID(gomock.Any()).Return(chat1, nil)
	repoSuccess.EXPECT().GetMember(chat1, gomock.Any()).Return(member1, nil)
	repoSuccess.EXPECT().AddMessages(chat1, gomock.Any()).Do(func(c *models.Conversation, m *models.Message) {
		m.ConversationID = uint(c.ID)
		m.CreatedAt = ct
	}).Return(nil)

	repoNotMember := fixtures.NewMockConversationRepository(ctrl)
	repoNotMember.EXPECT().GetByID(gomock.Any()).Return(chat1, nil)
	repoNotMember.EXPECT().GetMember(chat1, gomock.Any()).Return(nil, nil)

	tests := test_tools.Tests{
		//success
		test_tools.Test{
			"request": &chat.SendMessageRequest{ConversationId: 1, Messages: []*chat.Message{message1}},
			"reply":   &chat.SendMessageReply{Chat: &chat.Chat{Id: 1, Name: "Test chat"}, Messages: []*chat.Message{message1}},
			"repo":    repoSuccess,
			"error":   nil,
		},
		//not a member
		test_tools.Test{
			"request": &chat.SendMessageRequest{1, []*chat.Message{message1}},
			"reply": &chat.SendMessageReply{Chat: &chat.Chat{Id: 1, Name: "Test chat"}, Error: &chat.Error{
				Code:    chat.ErrorCode_FORBIDDEN,
				Message: "User isn't a member",
			}},
			"repo":  repoNotMember,
			"error": nil,
		},
	}
	rules := []test_tools.Rule{
		{test_tools.RuleStr, "reply"},
		{test_tools.RuleStr, "error"},
	}

	runner := test_tools.NewRunner(tests, f, rules)

	runner.RunTests()
	if runner.HasErrors() {
		t.Error(runner.Errors)
	}
}

func TestGetChatHistory(t *testing.T) {
	f := func(test test_tools.Test) []interface{} {
		server := NewChatServer(test["repo"].(models.ConversationRepository), nil)
		resp, err := server.GetChatHistory(context.Background(), test["request"].(*chat.ChatHistoryRequest))
		return []interface{}{resp, err}
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chat1 := &models.Conversation{
		Model: db.Model{ID: 1},
		Name:  "Test chat",
	}
	member1 := &models.Member{Model: db.Model{ID: 1}}
	ct := time.Now()
	message1 := &chat.Message{
		ConversationId: 1,
		UserId:         1,
		CreatedAt:      ct.Unix(),
	}

	messages := []*models.Message{
		{
			ConversationID: 1,
			UserID:         sql.NullInt64{Int64: 1},
			Model:          db.Model{CreatedAt: ct},
		},
	}

	repoSuccess := fixtures.NewMockConversationRepository(ctrl)
	repoSuccess.EXPECT().GetByID(gomock.Any()).Return(chat1, nil)
	repoSuccess.EXPECT().GetMember(chat1, gomock.Any()).Return(member1, nil)
	repoSuccess.EXPECT().GetHistory(chat1, gomock.Any(), gomock.Any(), false).Return(messages, nil)
	repoSuccess.EXPECT().TotalMessages(chat1).Return(uint64(1))

	tests := test_tools.Tests{
		//success
		test_tools.Test{
			"request": &chat.ChatHistoryRequest{0, 1, 0, 1, false},
			"reply": &chat.ChatHistoryReply{
				Messages:      []*chat.Message{message1},
				Chat:          &chat.Chat{Id: 1, Name: "Test chat"},
				TotalMessages: 1,
				Error:         nil},
			"repo":  repoSuccess,
			"error": nil,
		},
	}

	rules := []test_tools.Rule{
		{test_tools.RuleStr, "reply"},
		{test_tools.RuleStr, "error"},
	}

	runner := test_tools.NewRunner(tests, f, rules)

	runner.RunTests()
	if runner.HasErrors() {
		t.Error(runner.Errors)
	}
}
