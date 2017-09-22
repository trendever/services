package fetcher

import (
	"accountstore/client"
	"encoding/json"
	"errors"
	"fetcher/consts"
	"fetcher/models"
	"fmt"
	"instagram"
	"net/http"
	"proto/bot"
	"time"
	"utils/db"
	"utils/log"
	"utils/nats"
)

// @TODO delete expired requests and send errors notify with them

func processRequests(meta *client.AccountMeta) error {
	var requests []models.DirectRequest
	err := db.New().
		Where("user_id = ?", meta.Get().UserID).
		Where("reply = '' OR reply IS NULL").
		Order("id").Limit(100).
		Find(&requests).Error
	if err != nil {
		return fmt.Errorf("failed to load requests: %v", err)
	}

	for _, req := range requests {
		err = processRequest(meta, &req)
		if err != nil {
			return err
		}
	}
	return nil
}

func processRequest(meta *client.AccountMeta, req *models.DirectRequest) error {
	reply := bot.Notify{ReplyKey: req.ReplyKey, SourceId: meta.Get().UserID}
	if req.Data == "" && req.Kind != bot.MessageType_FetchThread {
		log.Warn("skipping empty message %v", req)
		err := db.New().Delete(req).Error
		if err != nil {
			return fmt.Errorf("failed to remove handled request from pending table: %v", err)
		}
		reply.Error = consts.EmptyData
	} else {
		ig, err := meta.Delayed()
		if err != nil {
			return err
		}
		switch req.Kind {
		case bot.MessageType_CreateThread:
			err = createThread(ig, req, &reply)
		case bot.MessageType_Text:
			err = sendText(ig, req, &reply)
		case bot.MessageType_Image:
			err = sendImage(ig, req, &reply)
		case bot.MessageType_MediaShare:
			err = shareMedia(ig, req, &reply)
		case bot.MessageType_ReplyComment:
			err = sendComment(ig, req, &reply)
		case bot.MessageType_FetchThread:
			err = fetchThread(meta, req, &reply)
		default:
			log.Errorf("request with unknown type %v ignored", req.Kind)
			return nil
		}
		if err != nil {
			return err
		}
	}

	err := nats.StanPublish(DirectNotifySubject, &reply)
	if err != nil {
		if req.ID != 0 {
			// we will send it later
			saveReply(req, &reply)
		}
		return fmt.Errorf("failed to send reply via stan: %v", err)
	}
	if req.ID == 0 {
		log.Warn("zero id in Request %+v", req)
		return nil
	}
	// @CHECK i can not see any real reason to save request logs. am i wrong?
	err = db.New().Delete(req).Error
	if err != nil {
		return fmt.Errorf("failed to remove handled request from pending table: %v", err)
	}
	return nil
}

func saveReply(req *models.DirectRequest, reply *bot.Notify) {
	bytes, err := json.Marshal(reply)
	if err != nil {
		log.Errorf("failed to marshal reply '%v': %v", reply, err)
		return
	}
	req.Reply = string(bytes)
	err = db.New().Save(req).Error
	if err != nil {
		log.Errorf("failed to save reply: %v", err)
	}
}

// resend replies for request that was performed while nats was unavailable
func resendRepliesLoop() {
	for range time.Tick(time.Minute) {
		var requests []models.DirectRequest
		err := db.New().
			Select("id, reply").
			Where("reply != '' AND reply IS NOT NULL").
			Order("id").Limit(1000).
			Find(&requests).Error
		if err != nil {
			log.Errorf("failed to load requests: %v", err)
			continue
		}
		if len(requests) != 0 {
			log.Warn("found %v request(s) with unsended reply, trying to resend", len(requests))
		}
		for _, req := range requests {
			resendReplyAndDelete(&req)
		}
	}
}

func resendReplyAndDelete(req *models.DirectRequest) {
	var reply bot.Notify
	err := json.Unmarshal([]byte(req.Reply), &reply)
	if err != nil {
		log.Errorf("failed to unmarshal reply '%v': %v", req.Reply, err)
		err := db.New().Delete(req).Error
		if err != nil {
			log.Errorf("failed to remove request with malformed saved reply: %v", err)
		}
		return
	}
	err = nats.StanPublish(DirectNotifySubject, &reply)
	if err != nil {
		// no need spam errors to log here, it will be processed again any way
		return
	}
	err = db.New().Delete(req).Error
	if err != nil {
		log.Errorf("failed to remove request with resended reply: %v", err)
	}
}

func createThread(ig *instagram.Instagram, req *models.DirectRequest, result *bot.Notify) error {
	bot, err := global.pubPool.GetRandom()
	if err != nil {
		return err
	}
	req.Participants = append(req.Participants, bot.UserID)
	return sendText(ig, req, result)
}

func sendText(ig *instagram.Instagram, req *models.DirectRequest, result *bot.Notify) error {
	var (
		messageID string
		err       error
	)
	switch {
	case req.ThreadID != "":
		result.ThreadId = req.ThreadID
		messageID, err = ig.BroadcastText(req.ThreadID, req.Data)

	case len(req.Participants) != 0:
		result.ThreadId, messageID, err = ig.SendText(req.Data, req.Participants...)

	default:
		result.Error = consts.BadDestination
		return nil
	}

	switch {
	case err == nil:
		result.Messages = []*bot.Message{
			{MessageId: messageID},
		}
	case err.Error() == "Thread does not exist":
		result.Error = consts.ThreadNotFound
		return nil
	default:
		return err
	}

	if req.Caption != "" {
		_, err := ig.DirectUpdateTitle(result.ThreadId, req.Caption)
		if err != nil {
			log.Errorf("set title for thread %v failed: %v", result.ThreadId, err)
		}
	}
	return nil
}

func shareMedia(ig *instagram.Instagram, req *models.DirectRequest, result *bot.Notify) error {
	if req.ThreadID == "" {
		result.Error = consts.BadDestination
		return nil
	}
	result.ThreadId = req.ThreadID
	messageID, err := ig.ShareMedia(req.ThreadID, req.Data)
	switch {
	case err == nil:
		result.Messages = []*bot.Message{
			{MessageId: messageID},
		}
	case err.Error() == "Thread does not exist":
		result.Error = consts.ThreadNotFound
	case err.Error() == "Media is not accessible":
		result.Error = consts.InaccessibleMedia
	default:
		return err
	}
	return nil
}

func sendImage(ig *instagram.Instagram, req *models.DirectRequest, result *bot.Notify) error {
	if req.ThreadID == "" {
		result.Error = consts.BadDestination
		return nil
	}
	result.ThreadId = req.ThreadID

	resp, err := http.Get(req.Data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 200:

	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		log.Errorf("failed to load image for request %v: bad status '%v(%v)'", req, resp.Status, resp.StatusCode)
		result.Error = consts.InvalidImage
		return nil
	default:
		return fmt.Errorf("failed to load image for request %v: bad status '%v(%v)'", req, resp.Status, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	switch contentType {
	case "", "application/octet-stream", "image/jpeg", "image/pjpeg":

	default:
		log.Errorf("unexpected content type '%v' for request %v", contentType, req)
		result.Error = consts.InvalidImage
		return nil
	}

	messageID, err := ig.SendPhoto(req.ThreadID, resp.Body)
	switch {
	case err == nil:
		result.Messages = []*bot.Message{
			{MessageId: messageID},
		}
	case err.Error() == "Thread does not exist":
		result.Error = consts.ThreadNotFound
	default:
		return err
	}
	return nil
}

func sendComment(ig *instagram.Instagram, req *models.DirectRequest, result *bot.Notify) error {
	if req.ThreadID == "" {
		result.Error = consts.BadDestination
		return nil
	}
	_, err := ig.CommentMedia(req.ThreadID, req.Data)
	switch {
	case err == nil:

	// @TODO find out reply for deleted post
	case err.Error() == "Sorry, this media has been deleted":
		result.Error = consts.InaccessibleMedia

	default:
		return err
	}
	return nil
}

func fetchThread(meta *client.AccountMeta, req *models.DirectRequest, result *bot.Notify) error {
	if req.ThreadID == "" {
		result.Error = consts.BadDestination
		return nil
	}
	result.ThreadId = req.ThreadID
	var err error
	ig := meta.Get()
	tmp := ig.ResponseLogging
	ig.ResponseLogging = true
	result.Messages, err = getEncodedThread(meta, req.ThreadID, req.Data)
	ig.ResponseLogging = tmp
	switch {
	case err == nil:
	case err.Error() == "Thread does not exist":
		result.Error = consts.ThreadNotFound
	default:
		return err
	}
	return nil
}

func MakeQuery(userID uint64, uri string) (reply string, err error) {
	ig := global.usersPool.Get(userID)
	if ig == nil {
		ig = global.pubPool.Get(userID)
		if ig == nil {
			return "", errors.New("user is unaviable")
		}
	}
	body, err := ig.RequestRaw(uri)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
