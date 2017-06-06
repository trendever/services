package wantit

import (
	"errors"
	"fmt"
	"proto/core"
	"utils/log"
	"utils/mandible"
	"utils/rpc"
	"wantit/api"
)

var (
	errorShopIsDeleted = errors.New("Shop is deleted; product will not be added")
)

func shopID(supplierID uint64) (uint64, error) {
	if supplierID == 0 {
		return 0, errors.New("zero supplierID")
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	// get or create shop
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
	switch resp := err.(type) {
	case nil:

	case *mandible.ImageResp:
		if resp.Status < 400 || resp.Status >= 500 {
			return nil, err
		}
		log.Warn("instagram user %v have invalid avatar", userInfo.User.Username)

	default:
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
