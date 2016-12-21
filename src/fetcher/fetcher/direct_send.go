package fetcher

import (
	"accountstore/client"
	"fetcher/models"
	"fmt"
	"proto/bot"
	"utils/db"
	"utils/log"
	"utils/nats"
)

// @TODO delete expired requests and send errors notify with them

func processRequests(meta *client.AccountMeta) error {
	var requests []models.DirectRequest
	err := db.New().Where("user_id = ?", meta.Get().UserID).Order("id").Limit(100).Find(&requests).Error
	if err != nil {
		return fmt.Errorf("failed to load requests: %v", err)
	}

	for _, req := range requests {
		switch req.Type {
		case models.SendMessageRequest:
			err = sendMessage(meta, &req)
		case models.CreateThreadRequest:
			err = createThread(meta, &req)
		default:
			log.Errorf("unknown request type %v", req.Type)
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func sendMessage(meta *client.AccountMeta, req *models.DirectRequest) error {
	ig, err := meta.Delayed()
	if err != nil {
		return err
	}
	reply := bot.DirectNotify{ReplyKey: req.ReplyKey}
	switch {
	case req.ThreadID != "":
		reply.ThreadId = req.ThreadID
		reply.MessageId, err = ig.BroadcastText(req.ThreadID, req.Text)
	case len(req.Participants) != 0:
		reply.ThreadId, reply.MessageId, err = ig.SendText(req.Text, req.Participants...)
	default:
		reply.Error = "destination is unspecified"
	}
	if err != nil {
		return err
	}

	if req.Caption != "" {
		_, err = ig.DirectUpdateTitle(reply.ThreadId, req.Caption)
		if err != nil {
			log.Errorf("set title for thread %v failed:", reply.ThreadId, err)
		}
	}

	err = nats.StanPublish(DirectNotifySubject, &reply)
	if err != nil {
		return fmt.Errorf("failed to send reply via stan: %v", err)
	}
	if req.ID == 0 {
		log.Warn("zero id in DirectRequest")
		return nil
	}
	// @CHECK i can not see any real reason to save request logs. am i wrong?
	err = db.New().Delete(req).Error
	if err != nil {
		return fmt.Errorf("failed to remove handled request from pending table: %v", err)
	}
	return nil
}

func createThread(meta *client.AccountMeta, req *models.DirectRequest) error {
	bot, err := global.pubPool.GetRandom()
	if err != nil {
		return err
	}
	req.Participants = append(req.Participants, bot.UserID)
	return sendMessage(meta, req)
}
