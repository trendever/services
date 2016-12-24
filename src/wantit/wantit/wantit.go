package wantit

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"wantit/api"
	"wantit/conf"

	"accountstore/client"
	"instagram"
	"proto/accountstore"
	"proto/bot"
	"proto/core"
	"utils/log"
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
	lastChecked = int64(0)
	pool        *client.AccountsPool
	settings    = conf.GetSettings()
	// @TODO products with big id will have more symbols
	codeRegexp     = regexp.MustCompile("t[a-z]+[0-9]{4}($|[^0-9])")
	avatarUploader = mandible.New(conf.GetSettings().MandibleURL)
)

// ResetLastChecked drops last checked id
func (svc *ProjectService) ResetLastChecked() error {
	return os.Remove(conf.GetSettings().LastCheckedFile)
}

// Run fetching
func (svc *ProjectService) Run() (err error) {

	rand.Seed(time.Now().Unix())
	api.Start()

	nats.Init(&settings.Nats, true)

	conn := rpc.Connect(settings.Instagram.StoreAddr)
	cli := accountstore.NewAccountStoreServiceClient(conn)
	instagram.DoResponseLogging = settings.Instagram.ResponseLogging
	pool, err = client.InitPoll(
		accountstore.Role_AuxPrivate, cli,
		nil, nil,
		&settings.Instagram.Settings,
	)
	if err != nil {
		return fmt.Errorf("failed to init acoounts pool: %v", err)
	}

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// we can safely start main loop even before registering apis
	// because every API request will take a connection from the connection pool
	// with getFreeApi()
	restoreLastChecked()
	go registerOrders()

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
	log.Debug("Loaded last checked id: %v", lastChecked)
}

func saveLastChecked() {
	ioutil.WriteFile(conf.GetSettings().LastCheckedFile, []byte(strconv.FormatInt(lastChecked, 10)), 0644)
}

func registerOrders() {
	timeout, _ := time.ParseDuration(settings.Instagram.TimeoutMin)
	loopStarted := time.Now()

	for {
		// make some delays in case loops runs too fast
		// startup delay is OK
		if time.Since(loopStarted) < time.Second {
			time.Sleep(timeout)
		}
		loopStarted = time.Now()
		log.Debug("Checking for new mention orders (last processed at %v)...", lastChecked)

		// Step #1: get new entries from fetcher
		res, err := retrieveActivities()

		if err != nil {
			log.Warn("RPC connection error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		log.Debug("... got %v results", len(res.Result))

		for _, mention := range res.Result {

			// add it
			retry, err := processPotentialOrder(mention.MediaId, mention)

			if err != nil {
				if retry {
					log.Error(err)
				} else {
					log.Warn(err.Error())
				}
			}

			if !retry {
				lastChecked = mention.Id
			} else {
				break
			}

		}
	}
}

func retrieveActivities() (*bot.RetrieveActivitiesReply, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	return api.FetcherClient.RetrieveActivities(ctx, &bot.RetrieveActivitiesRequest{
		Conds: []*bot.RetriveCond{
			{
				Role: bot.MentionedRole_Wantit,
				Type: []string{"mentioned", "direct"},
			},
			{
				Role: bot.MentionedRole_User,
				Type: []string{"commented", "direct"},
			},
		},
		AfterId: lastChecked,
		Limit:   100, //@CHECK this number
	})
}

// return arguments:
//   * retry bool. If true, this mention should be processed again lately
//   * error
func processPotentialOrder(mediaID string, mention *bot.Activity) (bool, error) {
	if mention.UserName == mention.MentionedUsername {
		return false, fmt.Errorf("Skipping self-mentioning activity (pk=%v)", mention.Pk)
	}

	// check if lead already registered
	if registered, err := isLeadRegistered(mention.Pk); registered {
		return false, fmt.Errorf("Skipping already added lead (pk=%v)", mention.Pk)
	} else if err != nil {
		return true, err
	}

	// get product media
	ig, err := pool.GetFree()
	if err != nil {
		return true, err
	}
	medias, err := ig.GetMedia(mediaID)
	if err != nil {
		if strings.Contains(err.Error(), "Media not found or unavailable") {
			return false, err
		}
		return true, err
	} else if len(medias.Items) != 1 {
		// deleted entry. @CHECK: anything else?
		return false, fmt.Errorf("Media not found (got result with %v items)", len(medias.Items))
	}

	productMedia := medias.Items[0]

	// get product via code
	var productID int64
	var deleted bool
	code, found := findProductCode(productMedia.Caption.Text)
	if found {
		productID, deleted, err = productCoreID(code)
		if err != nil {
			return true, err
		}
		if deleted {
			return false, errors.New("product was deleted")
		}
	}
	// there is no code at all or it's unregistred
	if !found || productID <= 0 {
		var retry bool
		productID, retry, err = saveProduct(mention)
		if retry {
			return true, fmt.Errorf("Temporarily unable to save product (%v)", err)
		}
		if err != nil {
			return true, err
		}
		if productID <= 0 {
			return false, errors.New("Could not save product: SaveTrend returned negative or zero productID")
		}
	}

	// check if self-mention
	if mention.UserName == productMedia.User.Username {
		log.Debug("Skipping order creation: @%v under own post (user=%v)", mention.MentionedUsername, productMedia.User.Username)
		return false, nil
	}

	// get customer core id
	customer, err := coreUser(mention.UserId, mention.UserName)
	if err != nil {
		return err != instagram.ErrorPageNotFound, err
	}

	if customer == nil {
		return false, fmt.Errorf("Core server returned nil customer for id %v", mention.UserId)
	}

	if customer.Seller {
		return false, fmt.Errorf("Skipping seller @wantit (for %v)", customer.InstagramUsername)
	}

	err = createOrder(mention, &productMedia, customer.Id, productID)
	if err != nil {
		return true, err
	}

	if !customer.Confirmed {
		err = notifyChat(mention)
		if err != nil {
			log.Errorf("Failed no reply in direct chat: %v", err)
		}
	}

	return false, nil
}

func saveProduct(mention *bot.Activity) (id int64, retry bool, err error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	log.Debug("Saving unknown product (activityId=%v)", mention.Id)

	res, err := api.SaveTrendClient.SaveProduct(ctx, mention)
	if err != nil {
		return -1, true, err
	}
	return res.Id, res.Retry, nil
}

func createOrder(mention *bot.Activity, media *instagram.MediaInfo, customerID, productID int64) error {

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	log.Debug("Creating new order (productId=%v)", productID)

	var source string
	switch mention.Type {
	case "commented":
		source = "comment"
	case "direct":
		source = "direct"
	default:
		source = "wantit"
	}

	_, err := api.LeadClient.CreateLead(ctx, &core.Lead{
		Source:        source,
		CustomerId:    customerID,
		ProductId:     int64(productID),
		Comment:       mention.Comment,
		InstagramPk:   mention.Pk,
		InstagramLink: fmt.Sprintf("https://www.instagram.com/p/%s/", media.Code),
	})

	return err
}

func notifyChat(mention *bot.Activity) error {
	//ctx, cancel := rpc.DefaultContext()
	//defer cancel()
	//
	//_, err := api.FetcherClient.SendDirect(ctx, &bot.SendDirectRequest{
	//	ActivityPk: mention.Pk,
	//	Text:       fmt.Sprintf(conf.GetSettings().DirectNotificationText, mention.UserName),
	//})
	//
	//return err

	// @TODO
	return nil
}

func findProductCode(comment string) (code string, found bool) {
	code = codeRegexp.FindString(comment)
	if code != "" {
		return code[:6], true
	}
	return "", false
}

// get core productId by mediaId
func productCoreID(code string) (id int64, deleted bool, err error) {

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := api.ProductClient.ReadProduct(ctx, &core.GetProductRequest{
		SearchBy:    &core.GetProductRequest_Code{code},
		WithDeleted: true,
	})

	if err != nil {
		return 0, false, err
	}

	return res.Id, res.Deleted, nil
}

// check if this lead alredy registered
func isLeadRegistered(commentPk string) (bool, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := api.LeadClient.ReadLead(ctx, &core.ReadLeadRequest{
		SearchBy: &core.ReadLeadRequest_InstagramPk{commentPk},
	})

	if err != nil {
		return false, err
	}
	return res.Id > 0, nil
}

func coreUser(instagramID uint64, instagramUsername string) (*core.User, error) {

	if instagramID == 0 || instagramUsername == "" {
		return nil, fmt.Errorf("zero instagram{ID/Name}in userId()")
	}

	// firstly, check if user exists
	user, err := findUser(instagramUsername)
	if err == nil && user != nil && user.Id > 0 {
		return user, nil
	} else if err != nil {
		return nil, err
	}

	// secondly, get this user profile
	ig, err := pool.GetFree()
	if err != nil {
		return nil, err
	}
	userInfo, err := ig.GetUserNameInfo(instagramID)
	if err != nil {
		return nil, err
	}

	avatarURL, _, err := avatarUploader.UploadImageByURL(userInfo.User.ProfilePicURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	// create user
	res, err := api.UserClient.FindOrCreateUser(ctx, &core.CreateUserRequest{
		User: &core.User{
			InstagramId:        uint64(instagramID),
			InstagramUsername:  userInfo.User.Username,
			InstagramFullname:  userInfo.User.FullName,
			InstagramAvatarUrl: userInfo.User.ProfilePicURL,
			InstagramCaption:   userInfo.User.Biography,
			Website:            userInfo.User.ExternalURL,
			AvatarUrl:          avatarURL,
		},
	})

	if err != nil {
		return nil, err
	}

	return res.User, nil
}

func findUser(instagramUsername string) (*core.User, error) {

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := api.UserClient.ReadUser(ctx, &core.ReadUserRequest{
		InstagramUsername: instagramUsername,
	})

	return res.User, err
}
