package fetcher

import (
	"accountstore/client"
	"fetcher/conf"
	"fmt"
	"proto/accountstore"
	"sync"
	"utils/log"
	"utils/rpc"
)

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
	global.msgChans[meta.Get().UserID] = msgChan
	global.Unlock()
	defer func() {
		global.Lock()
		ch := global.msgChans[meta.Get().UserID]
		if ch == msgChan {
			delete(global.msgChans, meta.Get().UserID)
		}
		global.Unlock()
	}()
	for {
		select {
		case <-stopChan:
			return
		case req := <-msgChan: // handle send direct message requests
			req.handle(meta)
		default: // nothing interesting; let's check feeds
			err := getActivity(meta)
			if err != nil {
				log.Errorf("failed to check instagram feed for user %v: %v", meta.Get().Username, err)
			}
			err = checkDirect(meta)
			if err != nil {
				log.Errorf("failed to check instagram direct for user %v: %v", meta.Get().Username, err)
			}
			err = parseOwnPosts(meta)
			if err != nil {
				log.Errorf("failed to check instagram own posts %v: %v", meta.Get().Username, err)
			}
		}
	}
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
