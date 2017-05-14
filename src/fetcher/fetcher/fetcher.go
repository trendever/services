package fetcher

import (
	"accountstore/client"
	"fetcher/conf"
	"fmt"
	"proto/accountstore"
	"utils/log"
	"utils/rpc"
)

var global = struct {
	pubPool   *client.AccountsPool
	usersPool *client.AccountsPool
}{}

// Start starts main fetching duty
func Start() error {
	settings := conf.GetSettings()

	conn := rpc.Connect(settings.Instagram.StoreAddr)
	cli := accountstore.NewAccountStoreServiceClient(conn)

	var err error
	global.pubPool, err = client.InitPoll(
		accountstore.Role_AuxPublic, cli,
		nil, pubWorker,
		&settings.Instagram.Settings,
	)
	if err != nil {
		return fmt.Errorf("failed to init acoounts pool: %v", err)
	}

	_, err = client.InitPoll(
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

	return nil
}

func primaryWorker(meta *client.AccountMeta, stopChan chan struct{}) {
	var step = 0
	for {
		select {
		case <-stopChan:
			return
		default:
			err := processRequests(meta)
			if err != nil {
				log.Warn("failed to process pending requests form user %v: %v", meta.Get().Username, err)
				continue
			}
			switch step {
			case 0:
				err := getActivity(meta)
				if err != nil {
					log.Errorf("failed to check instagram feed for user %v: %v", meta.Get().Username, err)
				}
			case 1:
				err := checkDirect(meta)
				if err != nil {
					log.Errorf("failed to check instagram direct for user %v: %v", meta.Get().Username, err)
				}
			case 2:
				err := parseOwnPosts(meta)
				if err != nil {
					log.Errorf("failed to check instagram own posts %v: %v", meta.Get().Username, err)
				}
			}
			step = (step + 1) % 3
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
