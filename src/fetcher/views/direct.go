package views

import (
	"fetcher/fetcher"
	"fetcher/models"
	"fmt"

	"golang.org/x/net/context"

	"proto/bot"
	"utils/db"
	"utils/log"
)

// SendDirect sends message to the chat (if not sent earlier)
func (s fetcherServer) SendDirect(ctx context.Context, in *bot.SendDirectRequest) (*bot.SendDirectReply, error) {

	// find related activity to get bot username
	var act models.Activity
	err := db.New().Where("pk = ?", in.ActivityPk).Find(&act).Error
	if err != nil {
		return nil, err
	}

	worker, err := fetcher.GetWorker(act.MentionedUsername)
	if err != nil {
		return nil, err
	}

	// do sending async
	go func() {
		var err error
		if act.ThreadID == "" { // create new thread
			err = sendDirectToNewChat(in, &act, worker)
		} else {
			err = sendDirectToChat(in, &act, worker)
		}

		if err != nil {
			log.Errorf("Could not create chat: %v", err)
		}
	}()

	return &bot.SendDirectReply{}, nil
}

func sendDirectToNewChat(req *bot.SendDirectRequest, act *models.Activity, worker *fetcher.Worker) error {
	res, err := worker.SendDirectMsgToUser(act.UserID, req.Text)
	if err != nil {
		return err
	}

	info, err := models.GetThreadInfo(res.ThreadID)
	if err != nil {
		return err
	}

	info.Notified = true
	err = info.Save()
	if err != nil {
		return nil
	}

	act.ThreadID = res.ThreadID
	return act.Save()
}

func sendDirectToChat(req *bot.SendDirectRequest, act *models.Activity, worker *fetcher.Worker) error {
	// find existing thread info
	info, err := models.GetThreadInfo(act.ThreadID)

	if info.Notified { // all ok; do nothing
		return nil
	}

	err = worker.SendDirectMsg(info.ThreadID, req.Text)
	if err != nil {
		return fmt.Errorf("Could not send shiet: %v", err)
	}

	// set notified
	// update only one column not to conflict with direct message crawling
	return db.New().
		Model(&models.ThreadInfo{}).
		Where("thread_id = ?", info.ThreadID).
		Update("notified", true).
		Error
}
