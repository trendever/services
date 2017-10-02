package wantit

import (
	"common/log"
	"errors"
	"fmt"
	"instagram"
	"proto/bot"
	"proto/core"
	"strings"
	"time"
	"utils/rpc"
	"wantit/api"
)

func processThreadOrder(mention *bot.Activity) (bool, error) {
	if mention.UserName == mention.MentionedUsername {
		return false, fmt.Errorf("Skipping self-mentioning activity (pk=%v)", mention.Pk)
	}

	if registered, err := isLeadRegistered(mention.Pk); registered {
		return false, fmt.Errorf("Skipping already added lead (pk=%v)", mention.Pk)
	} else if err != nil {
		return true, err
	}

	supplier, err := findUser(mention.MentionedUsername)
	if err != nil {
		return true, err
	}

	var product_code string = fmt.Sprintf("%v_help", supplier.Id)
	productID, deleted, err := productCoreID(product_code)
	if err != nil {
		return true, err
	}
	if deleted {
		return false, errors.New("product was deleted")
	}
	if productID <= 0 {
		var retry bool
		productID, retry, err = saveHelpProduct(mention, product_code, supplier.Id)
		if retry {
			return true, fmt.Errorf("Temporarily unable to save product (%v)", err)
		}
		if err != nil {
			return retry, err
		}
		if productID <= 0 {
			return false, errors.New("Could not save product: SaveTrend returned negative or zero productID")
		}
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

	err = createThreadOrder(mention, customer.Id, productID)
	if err != nil {
		return true, err
	}

	return false, nil

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
	ig, err := pool.GetFree(time.Minute)
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
		// this product belongs to someone else probably, so we will not have access to this thread
		// @TODO @CHECK may someone use codes to create multiple posts with same own product?
		mention.DirectThreadId = ""
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

	return false, nil
}

func createThreadOrder(mention *bot.Activity, customerID, productID int64) error {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	log.Debug("Creating new order (productId=%v)", productID)

	_, err := api.LeadClient.CreateLead(ctx, &core.Lead{
		Source:       "direct",
		DirectThread: mention.DirectThreadId,
		CustomerId:   customerID,
		ProductId:    int64(productID),
		Comment:      mention.Comment,
		InstagramPk:  mention.Pk,
	})

	return err
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
		Source:           source,
		DirectThread:     mention.DirectThreadId,
		CustomerId:       customerID,
		ProductId:        int64(productID),
		Comment:          mention.Comment,
		InstagramPk:      mention.Pk,
		InstagramLink:    fmt.Sprintf("https://www.instagram.com/p/%s/", media.Code),
		InstagramMediaId: mention.MediaId,
	})

	return err
}
