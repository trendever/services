package main

import (
	"golang.org/x/net/context"
	"instagram"
	"io/ioutil"
	"proto/checker"
	"regexp"
	"strconv"
	"strings"
	"time"
	"utils/db"
	"utils/log"
)

type CheckerServer struct {
	queryChan chan uint64
}

func NewCheckerServer() *CheckerServer {
	c := &CheckerServer{
		queryChan: make(chan uint64),
	}
	go c.loop()
	return c
}

func (c *CheckerServer) Check(ctx context.Context, in *checker.CheckRequest) (*checker.CheckReply, error) {
	go func(ids []uint64) {
		for _, id := range ids {
			c.queryChan <- id
		}
	}(in.Ids)
	return &checker.CheckReply{}, nil
}
func (c *CheckerServer) Stop() {
	close(c.queryChan)
}

// we don't need to load full user model, here is restricted version
type User struct {
	ID                 uint64 `gorm:"primaty_key"`
	UpdatedAt          time.Time
	Name               string
	InstagramID        uint64
	InstagramUsername  string
	InstagramAvatarURL string
	AvatarURL          string
}

func (u User) TableName() string {
	return "users_user"
}

func (s *CheckerServer) loop() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(settings.MinimalTickLen))
	lastChecked, err := loadLastChecked()
	if err != nil {
		log.Debug("failed to load last checked user id, starting from first")
	}
	for {
		select {
		case <-ticker.C:
			var users []User
			err := db.New().
				Where("id > ?", lastChecked).
				Where("instagram_username != ''").
				Where("deleted_at IS NULL").
				Limit(settings.RequestsPerTick).Order("id ASC").
				Find(&users).Error
			if err != nil {
				log.Errorf("failed to load users for update: %v", err)
			}
			for _, user := range users {
				checkUser(&user)
				lastChecked = user.ID
			}
			if uint64(len(users)) < settings.RequestsPerTick {
				lastChecked = 0
			}
			log.Debug("%v users have been checked", len(users))
			err = saveLastChecked(lastChecked)
			if err != nil {
				log.Errorf("failed to save last checked: %v", err)
			}
		case id, ok := <-s.queryChan:
			if !ok {
				return
			}
			user := User{ID: id}
			err := db.New().First(&user).Error
			if err != nil {
				log.Errorf("failed to load user for update: %v", err)
			}
			checkUser(&user)
			log.Debug("user %v updated by request", id)
		}
	}
}

var nameValidator = regexp.MustCompile("^(\\w|\\.)*$")

func checkUser(user *User) {
	if user.ID == 0 {
		return
	}
	var instagramInfo *instagram.SearchUserInfo
	updateMap := map[string]interface{}{}
	trimmed := strings.Trim(user.InstagramUsername, " \r\n\t")
	if trimmed != user.InstagramUsername {
		user.InstagramUsername = trimmed
		updateMap["instagram_username"] = trimmed
	}
	if nameValidator.MatchString(user.InstagramUsername) {
		candidates, err := Instagram.GetFree().SearchUsers(user.InstagramUsername)
		if err != nil {
			log.Errorf("failed to search user '%v' in instagram: %v", user.InstagramUsername, err)
			return
		}
		for i := range candidates.Users {
			if candidates.Users[i].Username == user.InstagramUsername {
				instagramInfo = &candidates.Users[i]
				break
			}
		}
	}
	// user not found
	if instagramInfo == nil {
		if user.Name == "" {
			updateMap["name"] = user.InstagramUsername
		}
		updateMap["instagram_username"] = ""
		updateMap["instagram_id"] = 0
	} else {
		if uint64(instagramInfo.Pk) != user.InstagramID {
			updateMap["instagram_id"] = instagramInfo.Pk
		}
		if instagramInfo.ProfilePicURL != user.InstagramAvatarURL {
			avatarURL, _, err := ImageUploader.UploadImageByURL(instagramInfo.ProfilePicURL)
			if err == nil {
				updateMap["instagram_avatar_url"] = instagramInfo.ProfilePicURL
				updateMap["avatar_url"] = avatarURL
			} else {
				log.Errorf("failed to upload new avatar for user %v: %v", user.ID, err)
			}
		}
	}
	if len(updateMap) != 0 {
		err := db.New().Model(&user).UpdateColumns(updateMap).Error
		if err != nil {
			log.Errorf("failed to update user %v: %v", user.ID, err)
		}
	}
}

// move it to utils?
func loadLastChecked() (uint64, error) {
	bytes, err := ioutil.ReadFile(settings.LastCheckedFile)
	if err != nil {
		return 0, err
	}
	res, err := strconv.ParseUint(string(bytes), 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func saveLastChecked(last uint64) error {
	return ioutil.WriteFile(settings.LastCheckedFile, []byte(strconv.FormatUint(last, 10)), 0644)
}
