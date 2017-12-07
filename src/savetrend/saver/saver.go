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

	"accountstore/client"
	"common/log"
	"instagram"
	"proto/accountstore"
	"proto/bot"
	"proto/core"
	"utils/mandible"
	"utils/nats"
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
	pool               *client.AccountsPool
	settings           = conf.GetSettings()
)

// ResetLastChecked drops last checked id
func (svc *ProjectService) ResetLastChecked() error {
	return os.Remove(conf.GetSettings().LastCheckedFile)
}

// Run fetching
func (svc *ProjectService) Run() (err error) {

	rand.Seed(time.Now().Unix())
	api.Start()

	settings := conf.GetSettings()
	instagram.ForceDebug = settings.Instagram.ResponseLogging

	nats.Init(&settings.Nats, true)

	conn := rpc.Connect(settings.Instagram.StoreAddr)
	cli := accountstore.NewAccountStoreServiceClient(conn)
	pool, err = client.InitPoll(
		accountstore.Role_AuxPrivate, cli,
		nil, nil,
		&settings.Instagram.Settings,
	)
	if err != nil {
		return fmt.Errorf("failed to init acoounts pool: %v", err)
	}

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

	// wait for terminating
	<-interrupt
	saveLastChecked()
	log.Warn("Cleanup and terminating...")
	client.StopAll()
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

func registerProducts() {
	timeout, _ := time.ParseDuration(settings.Instagram.TimeoutMin)
	loopStarted := time.Now()

	for {
		// make some delays in case loops runs too fast
		// startup delay is OK
		if time.Since(loopStarted) < time.Second {
			time.Sleep(timeout)
		}
		loopStarted = time.Now()
		// Step #1: get new entries from fetcher
		res, err := retrieveActivities()

		if err != nil {
			log.Warn("RPC connection error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if len(res.Result) > 0 {
			log.Debug("Got %v results since %v", len(res.Result), lastChecked)
		}

		for _, activity := range res.Result {
			log.Debug("processing activity %+v", activity)
			if _, retry, err := processProductMedia(activity.MediaId, activity); err != nil {
				if err != nil && retry {
					log.Debug("Retrying (%v)", err)
					break
				} else if err != nil {
					log.Debug("Skipping (%v)", err)
				}
			}

			// update last checked ID
			lastChecked = activity.Id
		}
		saveLastChecked()
	}
}

func retrieveActivities() (*bot.RetrieveActivitiesReply, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	return api.FetcherClient.RetrieveActivities(ctx, &bot.RetrieveActivitiesRequest{
		Conds: []*bot.RetriveCond{
			{
				Role: bot.MentionedRole_Savetrend,
				Type: []string{"mentioned", "direct"},
			},
			{
				Role: bot.MentionedRole_User,
				Type: []string{"ownfeed"},
			},
		},
		AfterId: lastChecked,
		Limit:   100, //@CHECK this number
	})
}

// processProductMedia returns id of product or error and retry flag
// returns:
//  * productID int64
//  * retry bool
//  * err error
func processProductMedia(mediaID string, mention *bot.Activity) (int64, bool, error) {
	mentionerID, _, err := userID(mention.UserId, mention.UserName)
	if err != nil {
		return -1, err != instagram.ErrorPageNotFound, fmt.Errorf("unable to get metioner id: %v", err)
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
	ig, err := pool.GetFree(time.Minute)
	if err != nil {
		return -1, true, err
	}
	medias, err := ig.GetMedia(mediaID)
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
		supplierInstagramID uint64
		supplierUsername    string
	)

	supplierInstagramID = productMedia.User.Pk
	supplierUsername = productMedia.User.Username

	supplierID, _, err := userID(supplierInstagramID, supplierUsername)
	if err != nil {
		return -1, true, fmt.Errorf("unable to get supplier id: %v", err)
	}
	shopID, err := shopID(uint64(supplierID))
	if err == errorShopIsDeleted {
		// ignore deleted shops
		return -1, false, err
	} else if err != nil {
		return -1, true, err
	}

	productID, retry, err := createProduct(&productMedia, shopID, mentionerID)
	if err != nil {
		return -1, retry, err
	}

	return productID, false, nil
}

func createProduct(media *instagram.MediaInfo, shopID, mentionerID uint64) (id int64, retry bool, err error) {
	var img *instagram.ImageCandidate
	switch media.MediaType {
	case instagram.MediaType_Image, instagram.MediaType_Video:
		img = media.ImageVersions2.Largest()
	case instagram.MediaType_Carousel:
		if len(media.CarouselMedia) == 0 {
			return -1, false, fmt.Errorf("Media %v has empty carousel", media.ID)
		}
		img = media.CarouselMedia[0].ImageVersions2.Largest()
	default:
		return -1, false, fmt.Errorf("Media %v has unsupported type", media.ID)
	}

	// @CHECK so what? why to not add product without image?
	if img == nil {
		return -1, false, fmt.Errorf("Media %v does not have parsable images", media.ID)
	}

	candidates, err := generateThumbnails(img.URL)
	switch resp := err.(type) {
	case nil:

	case *mandible.ImageResp:
		if resp.Status < 400 || resp.Status >= 500 {
			return -1, true, err
		}
		log.Warn("medai %v have invalid image", media.ID)

	default:
		return -1, true, err
	}

	request := &core.CreateProductRequest{Product: &core.Product{
		SupplierId:            int64(shopID),
		MentionedId:           int64(mentionerID),
		InstagramImageId:      media.ID,
		InstagramImageCaption: media.Caption.Text,
		InstagramLink:         fmt.Sprintf("https://www.instagram.com/p/%s/", media.Code),
		InstagramLikesCount:   int32(media.LikeCount),
		InstagramPublishedAt:  media.TakenAt,

		InstagramImages:      candidates,
		InstagramImageUrl:    img.URL,
		InstagramImageHeight: uint32(img.Height),
		InstagramImageWidth:  uint32(img.Width),
	}}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	log.Debug("Creating producto! Supplier=%v Mentioner=%v", shopID, mentionerID)

	res, err := api.ProductClient.CreateProduct(ctx, request)
	if err != nil {
		return -1, true, err
	}
	return res.Id, false, nil
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

// @TODO @REFACTOR move common wantit and savetrend logic in separate package or reorganise it in another way
// find core user with given instagramID; if not exists -- create one
func userID(instagramID uint64, instagramUsername string) (uint64, *core.User, error) {

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
	ig, err := pool.GetFree(time.Minute)
	if err != nil {
		return 0, nil, err
	}
	userInfo, err := ig.GetUserNameInfo(instagramID)
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
		switch resp := err.(type) {
		case nil:

		case *mandible.ImageResp:
			if resp.Status < 400 || resp.Status >= 500 {
				return 0, nil, err
			}
			log.Warn("instagram user %v have invalid avatar", userInfo.User.Username)

		default:
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
