package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"log"
	"proto/chat"
	"utils/rpc"
)

var cmdTest = &cobra.Command{
	Use:   "test",
	Short: "Test command",
	Run: func(cmd *cobra.Command, args []string) {
		conn := rpc.Connect(args[0])
		c := chat.NewChatServiceClient(conn)
		var resp interface{}
		var err error
		switch method {
		case "send":
			resp, err = send(c)
		case "create":
			resp, err = create(c)
		case "join":
			resp, err = join(c)
		case "history":
			resp, err = history(c)
		case "chats":
			resp, err = chats(c)
		case "readed":
			resp, err = readed(c)
		case "append":
			resp, err = msgAppend(c)
		default:
			log.Fatal("Unknown method")
		}
		j, _ := json.MarshalIndent(resp, "", "\t")
		fmt.Println(string(j), err)
	},
}

var (
	method, message, messageMime, userName   string
	userID, chatID, fromID, limit, messageID uint64
)

func init() {
	RootCmd.AddCommand(cmdTest)
	cmdTest.Flags().StringVarP(&method, "method", "m", "send", "method")
	cmdTest.Flags().StringVarP(&message, "message", "t", "", "message")
	cmdTest.Flags().Uint64VarP(&userID, "user", "u", 0, "user")
	cmdTest.Flags().StringVarP(&userName, "user_name", "n", "Guest", "user name")
	cmdTest.Flags().Uint64VarP(&chatID, "chat", "c", 0, "chat")
	cmdTest.Flags().Uint64VarP(&fromID, "from_id", "f", 0, "from id")
	cmdTest.Flags().Uint64VarP(&limit, "limit", "l", 0, "limit")
	cmdTest.Flags().Uint64VarP(&messageID, "msgid", "i", 0, "messageID")
	cmdTest.Flags().StringVarP(&messageMime, "mime", "M", "text/plain", "message mimetype")
}

func send(c chat.ChatServiceClient) (interface{}, error) {
	return c.SendNewMessage(context.Background(), &chat.SendMessageRequest{
		ConversationId: chatID,
		Messages: []*chat.Message{
			{
				UserId: userID,
				Parts: []*chat.MessagePart{
					{
						Content:  message,
						MimeType: messageMime,
					},
				},
			},
		},
	})
}

func msgAppend(c chat.ChatServiceClient) (interface{}, error) {
	return c.AppendMessage(context.Background(), &chat.AppendMessageRequest{
		MessageId: messageID,
		Parts: []*chat.MessagePart{
			{
				Content:  message,
				MimeType: messageMime,
			},
		},
	})
}

func create(c chat.ChatServiceClient) (interface{}, error) {
	return c.CreateChat(context.Background(), &chat.NewChatRequest{
		Chat: &chat.Chat{
			Name: message,
		},
	})
}

func join(c chat.ChatServiceClient) (interface{}, error) {
	return c.JoinChat(context.Background(), &chat.JoinChatRequest{
		ConversationId: chatID,
		Members: []*chat.Member{
			{
				UserId: userID,
				Role:   chat.MemberRole_CUSTOMER,
				Name:   userName,
			},
		},
	})
}

func history(c chat.ChatServiceClient) (interface{}, error) {
	return c.GetChatHistory(context.Background(), &chat.ChatHistoryRequest{
		ConversationId: chatID,
		UserId:         userID,
		FromMessageId:  fromID,
		Limit:          limit,
	})
}

func chats(c chat.ChatServiceClient) (interface{}, error) {
	return c.GetChats(context.Background(), &chat.ChatsRequest{
		UserId: userID,
		Id:     []uint64{chatID},
	})
}

func readed(c chat.ChatServiceClient) (interface{}, error) {
	return c.MarkAsReaded(context.Background(), &chat.MarkAsReadedRequest{
		UserId:         userID,
		ConversationId: chatID,
		MessageId:      fromID,
	})
}
