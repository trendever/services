package updater

import (
	"core/conf"
	"core/models"
	"fmt"
	"time"
	"utils/db"
	"utils/log"
)

const LastCheckedVarName = "last_checked_user"

type Updater struct {
	stopChan chan struct{}
}

func CreateAndRun() *Updater {
	u := &Updater{stopChan: make(chan struct{})}
	go u.loop()
	return u
}

func (u *Updater) Stop() {
	close(u.stopChan)
}

func (u *Updater) loop() {
	config := conf.GetSettings().Instagram.Updater
	ticker := time.NewTicker(time.Millisecond * time.Duration(config.MinimalTickLen))
	var lastChecked uint
	err := models.GetVar(LastCheckedVarName, &lastChecked)
	if err != nil {
		log.Debug("failed to load last checked user id, starting from first")
	}
	for {
		select {
		case <-ticker.C:
			var users []models.User
			err := db.New().Where("id > ?", lastChecked).Limit(config.RequestsPerTick).Order("id ASC").Find(&users).Error
			if err != nil {
				log.Error(fmt.Errorf("failed to load users for update: %v", err))
			}
			for _, user := range users {
				user.CheckInstagram()
				lastChecked = user.ID
			}
			if uint64(len(users)) < config.RequestsPerTick {
				lastChecked = 0
			}
			models.SetVar(LastCheckedVarName, lastChecked)
			log.Debug("%v users have been successfully checked with instagram updater", len(users))
		case <-u.stopChan:
			return
		}
	}
}
