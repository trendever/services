package views

import (
	"core/api"
	"core/db"
	"core/messager"
	"core/models"
	"core/telegram"
	"errors"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"utils/product_code"
)

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		core.RegisterProductServiceServer(s, productServer{
			usersProducts: models.NewUserProductsRepo(db.New()),
			searcher:      models.NewProductSearcher(db.New()),
		})
	})
}

type productServer struct {
	usersProducts models.UserProductsRepo
	searcher      models.ProductSearcher
}

func applyIDSearch(db *gorm.DB, request *core.GetProductRequest) (*gorm.DB, error) {

	// whether we want include deleted products
	if request.WithDeleted {
		db = db.Unscoped()
	}

	// here we are checking which field of oneof search_by we have
	switch {
	case request.GetId() > 0:
		return db.Where("id = ?", request.GetId()), nil
	case request.GetCode() != "":
		return db.Where("code = ?", request.GetCode()), nil //@TODO: check code for correctness?
	case request.GetMediaId() != "":
		return db.Where("instagram_image_id = ?", request.GetMediaId()), nil
	default:
		return nil, errors.New("Incorrect SearchBy")
	}
}

// get only one product by id/code
func (s productServer) GetProduct(ctx context.Context, request *core.GetProductRequest) (*core.ProductSearchResult, error) {

	query, err := applyIDSearch(models.DefaultProductQuery(), request)
	if err != nil {
		return nil, err
	}
	objects := models.Products{}
	res := query.
		Preload("LikedBy", "users_products.deleted_at IS NULL AND type = ?", "liked").
		Limit(1). // make sure only one is returned
		Find(&objects)

	if !res.RecordNotFound() && res.Error != nil {
		return nil, res.Error
	}

	return &core.ProductSearchResult{
		Result: objects.Encode(),
	}, nil
}

// get only one product by id/code
func (s productServer) ReadProduct(ctx context.Context, request *core.GetProductRequest) (*core.ProductReadResult, error) {
	query, err := applyIDSearch(db.New(), request)
	if err != nil {
		return nil, err
	}

	var ids []int64
	res := query.
		Limit(1). // make sure only one is returned
		Model(&models.Product{}).
		Pluck("id", &ids)

	var id int64

	if !res.RecordNotFound() && res.Error != nil {
		return nil, res.Error
	}

	if !res.RecordNotFound() && len(ids) == 1 {
		id = ids[0]
	}

	return &core.ProductReadResult{
		Id: id,
	}, nil
}

// search mutiple products
func (s productServer) SearchProducts(ctx context.Context, request *core.SearchProductRequest) (*core.ProductSearchResult, error) {
	var (
		objects models.Products
		err     error
		filter  = &models.ProductFilter{}
	)

	err = filter.Decode(request)

	if err != nil {
		return nil, err
	}

	ids, err := s.searcher.Search(*filter)

	if err != nil {
		return nil, err
	}

	objects, err = models.GetProductsByIDs(ids, filter.Direction)

	if err != nil {
		return nil, err
	}

	results := objects.Encode()

	return &core.ProductSearchResult{
		Result: results,
	}, nil
}

func (s productServer) CreateProduct(ctx context.Context, request *core.CreateProductRequest) (*core.CreateProductResult, error) {

	product := models.Product{}.Decode(request.Product)

	if len(product.Items) == 0 {
		// create empty product item. really important, because
		//  it allows to buy (from instagram) even a product, which was not processed by a manager)
		product.Items = []models.ProductItem{
			{},
		}
	}

	err := db.New().Create(&product).Error

	if err != nil {
		return nil, err
	}
	//add product to user profile
	s.usersProducts.Like(product.ID, product.MentionedByID)

	if product.Code == "" {
		product.Code = product_code.GenCode(int64(product.ID))
		db.New().Save(&product)
	}

	if product.MentionedBy.ID == 0 && product.MentionedByID > 0 {
		if user, err := models.GetUserByID(product.MentionedByID); err == nil {
			product.MentionedBy = *user
		}
	}

	if product.Shop.ID == 0 && product.ShopID > 0 {
		if shop, err := models.GetShopByID(product.ShopID); err == nil {
			product.Shop = *shop
		}
	}

	go telegram.NotifyProductCreated(&product)
	go messager.Publish("core.product.new", product.Encode())

	return &core.CreateProductResult{
		Id:   int64(product.ID),
		Code: product.Code,
	}, nil
}

func (s productServer) LikeProduct(_ context.Context, req *core.LikeProductRequest) (reply *core.LikeProductReply, err error) {
	reply = &core.LikeProductReply{}

	if req.UserId == 0 || req.ProductId == 0 {
		return nil, errors.New("user_id and product_id are required")
	}
	user, err := models.GetUserByID(uint(req.UserId))

	if err != nil {
		return
	}

	product, err := models.GetProductByID(req.ProductId)

	if err != nil {
		return
	}

	switch {
	case req.Like == true:
		err = s.usersProducts.Like(product.ID, user.ID)
	case req.Like == false:
		err = s.usersProducts.Unlike(product.ID, user.ID)
	}

	db.New().Model(product).
		Preload("LikedBy", "users_products.deleted_at IS NULL AND type = ?", "liked").
		Find(product, product.ID)

	if err == nil {
		go api.Publish("core.product.update", product.Encode())
	}

	return
}

func (s productServer) GetSpecialProducts(_ context.Context, _ *core.GetSpecialProductsRequest) (*core.GetSpecialProductsReply, error) {
	var list []*core.SpecialProductInfo
	res := db.New().
		Select("DISTINCT p.id, p.title").
		Table("settings_templates_chat t").
		Joins("INNER JOIN products_product p ON t.product_id = p.id").
		Scan(&list)
	if res.RecordNotFound() {
		return &core.GetSpecialProductsReply{}, nil
	}
	if res.Error != nil {
		return &core.GetSpecialProductsReply{Err: res.Error.Error()}, nil
	}
	return &core.GetSpecialProductsReply{List: list}, nil
}
