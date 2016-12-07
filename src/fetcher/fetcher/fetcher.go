package fetcher

import (
	"accountstore/client"
	"errors"
	"fetcher/conf"
	"fmt"
	"proto/accountstore"
	"sync"
	"utils/log"
	"utils/rpc"
)

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

var global = struct {
	sync.RWMutex
	pubPool   *client.AccountsPool
	usersPool *client.AccountsPool
	msgChans  map[uint64]chan sendRequest
}{
	msgChans: map[uint64]chan sendRequest{},
}

// Start starts main fetching duty
func Start() error {
	settings := conf.GetSettings()

	conn := rpc.Connect(settings.Instagram.StoreAddr)
	cli := accountstore.NewAccountStoreServiceClient(conn)

	_, err := client.InitPoll(
		accountstore.Role_Wantit, cli,
		nil, primaryWorker,
		&settings.Instagram.Settings,
	)
	if err != nil {
		return fmt.Errorf("failed to init acoounts pool: %v", err)
	}

	_, err = client.InitPoll(
		accountstore.Role_Savetrend, cli,
		nil, primaryWorker,
		&settings.Instagram.Settings,
	)
	if err != nil {
		return fmt.Errorf("failed to init acoounts pool: %v", err)
	}

	global.usersPool, err = client.InitPoll(
		accountstore.Role_User, cli,
		nil, primaryWorker,
		&settings.Instagram.Settings,
	)
	if err != nil {
		return fmt.Errorf("failed to init acoounts pool: %v", err)
	}

	global.pubPool, err = client.InitPoll(
		accountstore.Role_AuxPublic, cli,
		nil, pubWorker,
		&settings.Instagram.Settings,
	)
	if err != nil {
		return fmt.Errorf("failed to init acoounts pool: %v", err)
	}

	return nil
}

func primaryWorker(meta *client.AccountMeta, stopChan chan struct{}) {
	msgChan := make(chan sendRequest)
	global.Lock()
	global.msgChans[meta.Get().UserNameID] = msgChan
	global.Unlock()
	defer func() {
		global.Lock()
		ch := global.msgChans[meta.Get().UserNameID]
		if ch == msgChan {
			delete(global.msgChans, meta.Get().UserNameID)
		}
		global.Unlock()
	}()
	for {
		select {
		case <-stopChan:
			return
		case req := <-msgChan:
			ig, err := meta.Delayed()
			if err != nil {
				req.reply <- sendReply{error: err}
				continue
			}
			if req.threadID != "" {
				msgID, err := ig.BroadcastText(req.threadID, req.text)
				req.reply <- sendReply{msgID: msgID, threadID: req.threadID, error: err}
				continue
			}
			if req.receiverID != 0 {
				tid, err := ig.SendText(req.text, req.receiverID)
				req.reply <- sendReply{threadID: tid, error: err}
				continue
			}
			req.reply <- sendReply{error: errors.New("destination is unspecified")}
		default:
			err := getActivity(meta)
			if err != nil {
				log.Errorf("failed to check instagram feed for user %v: %v", meta.Get().Username, err)
			}
			err = checkDirect(meta)
			if err != nil {
				log.Errorf("failed to check instagram direct for user %v: %v", meta.Get().Username, err)
			}
		}
	}
}

func SendDirect(senderID, receiverID uint64, threadID, text string) (msgID string, err error) {
	global.RLock()
	ch, ok := global.msgChans[senderID]
	global.RUnlock()
	if !ok {
		return "", errors.New("sender not found")
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

func pubWorker(meta *client.AccountMeta, stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		default:
			err := leaveAllThreads(meta)
			if err != nil {
				log.Errorf("pub bot: failed to leave threads: %v", err)
			}
		}
	}
}
