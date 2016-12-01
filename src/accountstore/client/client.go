package client

import (
	"errors"
	"fmt"
	"instagram"
	"math/rand"
	"proto/accountstore"
	"sync"
	"time"
	"utils/log"
	"utils/nats"
	"utils/rpc"
	"utils/stopper"
)

func init() {
	nats.StanSubscribe(&nats.StanSubscription{
		Subject:        "accounts.notify",
		DecodedHandler: handleNotify,
	})
}

var global struct {
	sync.RWMutex
	pools []*AccountsPool
}

type AccountsPool struct {
	sync.RWMutex
	role     accountstore.Role
	idMap    map[uint64]*AccountMeta
	storeCli accountstore.AccountStoreServiceClient
	ready    chan *instagram.Instagram
	timeout  struct {
		min int
		max int
	}
	individualWorker func(acc *AccountMeta, stopChan chan struct{})
	// global stopper
	stopper *stopper.Stopper
}

type AccountMeta struct {
	ig    *instagram.Instagram
	pool  *AccountsPool
	ready chan *instagram.Instagram
	// stopper for individual workers
	stopper *stopper.Stopper
}

func (acc *AccountMeta) Get() (*instagram.Instagram, error) {
	select {
	case ig := <-acc.ready:
		return ig, nil
	case <-acc.stopper.Chan():
		return nil, errors.New("account is stopped")
	}
}

type Settings struct {
	TimeoutMin string
	TimeoutMax string
}

func InitPoll(
	role accountstore.Role,
	storeCli accountstore.AccountStoreServiceClient,
	poolWorker func(pool *AccountsPool, stopChan chan struct{}),
	individualWorker func(acc *AccountMeta, stopChan chan struct{}),
	settings Settings,
) (*AccountsPool, error) {

	min, err := time.ParseDuration(settings.TimeoutMin)
	if err != nil {
		return nil, err
	}
	max, err := time.ParseDuration(settings.TimeoutMax)
	if err != nil {
		return nil, err
	}

	pool := &AccountsPool{
		idMap:            make(map[uint64]*AccountMeta),
		storeCli:         storeCli,
		individualWorker: individualWorker,
		stopper:          stopper.NewStopper(),
		timeout: struct {
			min int
			max int
		}{min: int(min), max: int(max)},
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := storeCli.Search(ctx, &accountstore.SearchRequest{Roles: []accountstore.Role{role}})
	if err != nil {
		return nil, fmt.Errorf("failed to load instagram accounts: %v", err)
	}

	for _, acc := range res.Accounts {
		ig, err := instagram.Restore(acc.Cookie, "")
		if err != nil {
			log.Errorf("fialed to restore account %v: %v", acc.InstagramUsername, err)
			continue
		}
		pool.addAcc(ig)
	}

	if len(pool.idMap) == 0 {
		log.Warn("zero accounts provided with role %v", role)
	}

	if poolWorker != nil {
		go func() {
			poolWorker(pool, pool.stopper.Chan())
		}()
	}

	global.Lock()
	global.pools = append(global.pools, pool)
	global.Unlock()

	return pool, nil
}

func (pool *AccountsPool) randomTimeout() {
	rndTimeout := pool.timeout.min + rand.Intn(pool.timeout.max-pool.timeout.min)
	time.Sleep(time.Duration(rndTimeout))
}

func (pool *AccountsPool) Get(id uint64) (acc *instagram.Instagram, found bool) {
	pool.RLock()
	info, found := pool.idMap[id]
	pool.RUnlock()
	return info.ig, found
}

func (pool *AccountsPool) GetFree() (*instagram.Instagram, error) {
	for {
		select {
		case ig := <-pool.ready:
			if ig.LoggedIn {
				return ig, nil
			}
			pool.Invalidate(ig.UserNameID)
		case <-pool.stopper.Chan():
			return nil, errors.New("pool is stopped")
		}
	}
}

func (pool *AccountsPool) Invalidate(id uint64) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	_, err := pool.storeCli.MarkInvalid(ctx, &accountstore.MarkInvalidRequest{InstagramId: id})
	if err != nil {
		log.Errorf("failed to invalidate account %v: %v", id, err)
		return
	}

	pool.Lock()
	defer pool.Unlock()

	info, ok := pool.idMap[id]
	if !ok {
		log.Errorf("can not remove unknown account %v", id)
		return
	}
	info.stopper.Stop()
	delete(pool.idMap, id)
}

func (pool *AccountsPool) update(acc *accountstore.Account) {
	ig, err := instagram.Restore(acc.Cookie, "")
	if err != nil {
		log.Errorf("fialed to restore account %v: %v", acc.InstagramUsername, err)
		return
	}

	pool.Lock()
	defer pool.Unlock()

	if !acc.Valid {
		info, ok := pool.idMap[ig.UserNameID]
		if !ok {
			return
		}
		pool.delAcc(info)
		return
	}

	info, ok := pool.idMap[ig.UserNameID]
	// we have this account already
	if ok {
		if acc.Role != pool.role {
			pool.delAcc(info)
			return
		}
		// update data
		*info.ig = *ig
		// @CHECK do we need to restart worker?
		return
	}

	if acc.Role == pool.role {
		pool.addAcc(ig)
	}
}

// adds account to pool and starts individualWorker(if any),
// pool should be already locked on higher level
func (pool *AccountsPool) addAcc(ig *instagram.Instagram) {
	data := &AccountMeta{ig: ig, stopper: stopper.NewStopper()}
	pool.idMap[ig.UserNameID] = data

	go func() {
		for {
			select {
			case pool.ready <- ig:
				pool.randomTimeout()
			case data.ready <- ig:
				pool.randomTimeout()
			case <-data.stopper.Chan():
				return
			}
		}
	}()

	if pool.individualWorker != nil {
		go func() {
			pool.individualWorker(data, data.stopper.Chan())
		}()
	}
}

// removes account from pool and sends stop signal to related gorutines,
// pool should be already locked on higher level
func (pool *AccountsPool) delAcc(acc *AccountMeta) {
	acc.stopper.Stop()
	delete(pool.idMap, acc.ig.UserNameID)
}

func (pool *AccountsPool) Stop() {
	pool.Lock()
	for _, acc := range pool.idMap {
		pool.delAcc(acc)
	}
	pool.stopper.Stop()
	pool.Unlock()

	global.Lock()
	for k, v := range global.pools {
		if v == pool {
			global.pools[k] = global.pools[len(global.pools)-1]
			global.pools[len(global.pools)-1] = nil
			global.pools = global.pools[:len(global.pools)-1]
		}
	}
	global.Unlock()
}

func handleNotify(acc *accountstore.Account) bool {
	global.Lock()
	for _, pool := range global.pools {
		pool.update(acc)
	}
	global.Unlock()
	return true
}
