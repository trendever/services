package saver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"savetrend/api"
	"savetrend/conf"

	"instagram"
	"proto/bot"
	"proto/core"
	"utils/log"
	"utils/rpc"
)

type textField struct {
	userName string
	textType string
	comment  string
}

// ProjectService struct
type ProjectService struct{}

var (
	lastChecked        = int64(0)
	errorAlreadyAdded  = errors.New("Product already exists")
	errorShopIsDeleted = errors.New("Shop is deleted; product will not be added")
	pool               *instagram.Pool
	settings           = conf.GetSettings()
)

// ResetLastChecked drops last checked id
func (svc *ProjectService) ResetLastChecked() error {
	return os.Remove(conf.GetSettings().LastCheckedFile)
}

// Run fetching
func (svc *ProjectService) Run() error {

	rand.Seed(time.Now().Unix())
	api.Start()

	srv := rpc.Serve(settings.Rpc)
	bot.RegisterSaveTrendServiceServer(srv, NewSaveServer())

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// we can safely start main loop even before registering apis
	// because every API request will take a connection from the connection pool
	// with getFreeApi()
	restoreLastChecked()
	go registerProducts()

	err := registerApis()
	if err != nil {
		return err
	}

	// wait for terminating
	<-interrupt
	saveLastChecked()
	log.Warn("Cleanup and terminating...")
	os.Exit(0)

	return nil
}

func restoreLastChecked() {
	bytes, err := ioutil.ReadFile(conf.GetSettings().LastCheckedFile)
	if err != nil {
		log.Error(err)
		return
	}

	res, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		log.Error(err)
		return
	}

	lastChecked = res
	log.Info("Loaded last checked id: %v", lastChecked)
}

func saveLastChecked() {
	ioutil.WriteFile(conf.GetSettings().LastCheckedFile, []byte(strconv.FormatInt(lastChecked, 10)), 0644)
}

func registerApis() error {

	settings := conf.GetSettings()

	pool = instagram.NewPool(&instagram.PoolSettings{
		TimeoutMin:     settings.Instagram.TimeoutMin,
		TimeoutMax:     settings.Instagram.TimeoutMax,
		ReloginTimeout: settings.Instagram.ReloginTimeout,
	})

	// open connection and append connections pool
	for _, user := range conf.GetSettings().Instagram.Users {

		api, err := instagram.NewInstagram(user.Username, user.Password)
		if err != nil {
			return err
		}

		pool.Add(api)
	}

	return nil
}

func registerProducts() {

	for {
		log.Debug("Checking for new products (last checked at %v)", lastChecked)

		// Step #1: get new entries from fetcher
		res, err := retrieveActivities()

		if err != nil {
			log.Warn("RPC connection error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for _, mention := range res.Result {
			if _, retry, err := processProductMedia(mention.MediaId, mention); err != nil {
				if err != nil && retry {
					log.Debug("Retrying (%v)", err)
					break
				} else if err != nil {
					log.Debug("Skipping (%v)", err)
				}
			}

			// update last checked ID
			lastChecked = mention.Id
		}

		time.Sleep(time.Millisecond * time.Duration(settings.Instagram.PollTimeout))
	}
}

func retrieveActivities() (*bot.RetrieveActivitiesReply, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	return api.FetcherClient.RetrieveActivities(ctx, &bot.RetrieveActivitiesRequest{
		AfterId:     lastChecked,
		Type:        []string{"mentioned", "direct"},
		MentionName: conf.GetSettings().Instagram.TrendUser,
		Limit:       100, //@CHECK this number
	})
}

// processProductMedia returns id of product or error and retry flag
// returns:
//  * productID int64
//  * retry bool
//  * err error
func processProductMedia(mediaID string, mention *bot.Activity) (int64, bool, error) {

	mentionerID, mentioner, err := userID(mention.UserId, mention.UserName)
	if err != nil {
		return -1, true, err
	}

	productID, deleted, err := productExists(mediaID) //@TODO: batch check for existence
	if err != nil {
		return -1, true, err
	}
	if deleted {
		return -1, false, errors.New("product was deleted")
	}
	if productID > 0 {
		go func() {
			ctx, cancel := rpc.DefaultContext()
			defer cancel()
			//Product already exists, but we want to add it to user trends
			api.ProductClient.LikeProduct(ctx, &core.LikeProductRequest{
				UserId:    uint64(mentionerID),
				ProductId: uint64(productID),
				Like:      true,
			})
		}()
		return productID, false, errorAlreadyAdded
	}

	// read media info
	medias, err := pool.GetFree().GetMedia(mediaID)
	if err != nil {
		if strings.Contains(err.Error(), "Media not found or unavailable") {
			return -1, false, err
		}
		return -1, true, fmt.Errorf("failed to load media '%v': %v", mediaID, err)
	} else if len(medias.Items) != 1 {
		// this seems not no happen normally; so put Warning here
		err = fmt.Errorf("Media (%v) not found (got result with %v items)", mediaID, len(medias.Items))
		log.Warn(err.Error())
		return -1, false, err
	}

	var (
		productMedia        = medias.Items[0]
		supplierInstagramID int64
		supplierUsername    string
	)

	supplierInstagramID = productMedia.User.Pk
	supplierUsername = productMedia.User.Username

	supplierID, _, err := userID(supplierInstagramID, supplierUsername)
	if err != nil {
		return -1, true, err
	}
	shopID, err := shopID(uint64(supplierID))
	if err == errorShopIsDeleted {
		// ignore deleted shops
		return -1, false, err
	} else if err != nil {
		return -1, true, err
	}

	productID, err = createProduct(mediaID, &productMedia, shopID, mentionerID)
	if err != nil {
		return -1, true, err
	}

	if !mentioner.Confirmed {
		err = notifyChat(mention)
		if err != nil {
			log.Errorf("Failed no reply in direct chat: %v", err)
		}
	}

	return productID, false, nil
}

func createProduct(mediaID string, media *instagram.MediaInfo, shopID, mentionerID uint64) (id int64, err error) {

	if len(media.ImageVersions2.Candidates) < 1 {
		return -1, errors.New("Product media has no images!")
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	log.Debug("Creating producto! Supplier=%v Mentioner=%v", shopID, mentionerID)

	img := media.ImageVersions2.Candidates[0]

	candidates, err := generateThumbnails(img.URL)
	if err != nil {
		return -1, err
	}

	request := &core.CreateProductRequest{Product: &core.Product{
		SupplierId:            int64(shopID),
		MentionedId:           int64(mentionerID),
		InstagramImageId:      mediaID,
		InstagramImageCaption: media.Caption.Text,
		InstagramLink:         fmt.Sprintf("https://www.instagram.com/p/%s/", media.Code),
		InstagramLikesCount:   int32(media.LikeCount),
		InstagramPublishedAt:  media.TakenAt,

		InstagramImages:      candidates,
		InstagramImageUrl:    img.URL,
		InstagramImageHeight: uint32(img.Height),
		InstagramImageWidth:  uint32(img.Width),
	}}

	res, err := api.ProductClient.CreateProduct(ctx, request)
	if err != nil {
		return -1, err
	}
	return res.Id, nil
}

// check if product with this mediaId present.
func productExists(mediaID string) (id int64, deleted bool, err error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := api.ProductClient.ReadProduct(ctx, &core.GetProductRequest{
		SearchBy:    &core.GetProductRequest_MediaId{MediaId: mediaID},
		WithDeleted: true,
	})

	if err != nil {
		return 0, false, err
	}

	return res.Id, res.Deleted, nil
}

// find core user with given instagramID; if not exists -- create one
func userID(instagramID int64, instagramUsername string) (uint64, *core.User, error) {

	if instagramID == 0 {
		return 0, nil, errors.New("zero instagramId in userId()")
	}

	// firstly, check if user exists
	user, err := findUser(instagramUsername)
	if err == nil && user != nil && user.Id > 0 {
		return uint64(user.Id), user, nil
	} else if err != nil {
		return 0, nil, err
	}

	// secondly, get this user profile
	userInfo, err := pool.GetFree().GetUserNameInfo(instagramID)
	if err != nil {
		return 0, nil, err
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := api.UserClient.ReadUser(ctx, &core.ReadUserRequest{
		InstagramId: uint64(instagramID),
	})

	if err != nil {
		return 0, nil, err
	}

	if res.User.Id == 0 {
		// create user

		// upload avatar
		avatarURL, err := uploadAvatar(userInfo.User.ProfilePicURL)
		if err != nil {
			return 0, nil, err
		}

		// do create
		res, err = api.UserClient.FindOrCreateUser(ctx, &core.CreateUserRequest{
			User: &core.User{
				InstagramId:        uint64(instagramID),
				InstagramUsername:  userInfo.User.Username,
				Name:               userInfo.User.Username,
				InstagramFullname:  userInfo.User.FullName,
				InstagramAvatarUrl: userInfo.User.ProfilePicURL,
				InstagramCaption:   userInfo.User.Biography,
				Website:            userInfo.User.ExternalURL,
				AvatarUrl:          avatarURL,
			},
		})

		if err != nil {
			return 0, nil, err
		}
	}

	return uint64(res.User.Id), res.User, nil
}

// find user; returns rpc err and positive id if found
func findUser(instagramUsername string) (*core.User, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := api.UserClient.ReadUser(ctx, &core.ReadUserRequest{
		InstagramUsername: instagramUsername,
	})

	if res == nil {
		return nil, err
	}
	return res.User, err
}

// finds an exiting instagram shop for supplier; if not exists -- creates one
func shopID(supplierID uint64) (uint64, error) {
	if supplierID == 0 {
		return 0, errors.New("zero supplierID")
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	// create shop
	res, err := api.ShopClient.FindOrCreateShopForSupplier(
		ctx, &core.FindOrCreateShopForSupplierRequest{SupplierId: supplierID},
	)
	if err != nil {
		return 0, fmt.Errorf("RPC error: %v", err)
	}
	if res.Error != "" {
		return 0, errors.New(res.Error)
	}
	if res.Deleted {
		return 0, errorShopIsDeleted
	}

	return res.ShopId, nil
}

func notifyChat(mention *bot.Activity) error {

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	_, err := api.FetcherClient.SendDirect(ctx, &bot.SendDirectRequest{
		ActivityPk: mention.Pk,
		Text:       fmt.Sprintf(conf.GetSettings().DirectNotificationText, mention.UserName),
	})

	return err
}
