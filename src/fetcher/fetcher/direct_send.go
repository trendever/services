package fetcher

import (
	"accountstore/client"
	"errors"
)

var BadDestinationError = errors.New("destination is unspecified")

type sendReply struct {
	msgID    string
	threadID string
	error    error
}

type sendRequest struct {
	receiverID uint64
	threadID   string
	text       string
	reply      chan sendReply
}

// async send message request handler; reply via provided chan
func (req *sendRequest) handle(meta *client.AccountMeta) {
	ig, err := meta.Delayed()
	if err != nil {
		req.reply <- sendReply{error: err}
		return
	}
	if req.threadID != "" {
		msgID, err := ig.BroadcastText(req.threadID, req.text)
		req.reply <- sendReply{msgID: msgID, threadID: req.threadID, error: err}
		return
	}
	if req.receiverID != 0 {
		tid, mid, err := ig.SendText(req.text, req.receiverID)
		req.reply <- sendReply{threadID: tid, msgID: mid, error: err}
		return
	}
	req.reply <- sendReply{error: BadDestinationError}
}

func SendDirect(senderID, receiverID uint64, threadID, text string) (msgID string, err error) {
	global.RLock()
	ch, ok := global.msgChans[senderID]
	global.RUnlock()
	if !ok {
		return "", AccountUnavailable
	}
	replyChan := make(chan sendReply)
	ch <- sendRequest{
		threadID:   threadID,
		receiverID: receiverID,
		text:       text,
		reply:      replyChan,
	}
	reply := <-replyChan
	return reply.msgID, reply.error
}
