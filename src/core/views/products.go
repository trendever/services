package views

import (
	"core/api"
	"core/models"
	"errors"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"utils/db"
	"utils/log"
	"utils/nats"
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
		db = db.Unscoped().Order("deleted_at IS NOT NULL")
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

// GetProduct fetches whole product by id/code
func (s productServer) GetProduct(ctx context.Context, request *core.GetProductRequest) (*core.ProductSearchResult, error) {

	query, err := applyIDSearch(models.DefaultProductQuery(), request)
	if err != nil {
		return nil, err
	}
	objects := models.Products{}
	res := query.
		Preload("LikedBy", "users_products.deleted_at IS NULL AND type = ?", "liked").
		Preload("Shop").
		Preload("Shop.Supplier").
		Limit(1). // make sure only one is returned
		Find(&objects)

	if !res.RecordNotFound() && res.Error != nil {
		return nil, res.Error
	}

	return &core.ProductSearchResult{
		Result: objects.Encode(),
	}, nil
}

// ReadProduct checks product existence by id/code
func (s productServer) ReadProduct(ctx context.Context, request *core.GetProductRequest) (*core.ProductReadResult, error) {
	query, err := applyIDSearch(db.New(), request)
	if err != nil {
		return nil, err
	}

	var reply core.ProductReadResult
	res := query.
		Limit(1). // make sure only one is returned
		Model(&models.Product{}).
		Select("id, deleted_at IS NOT NULL AS deleted").
		Scan(&reply)

	if !res.RecordNotFound() && res.Error != nil {
		return nil, res.Error
	}

	return &reply, nil
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

	go models.GetNotifier().NotifyAboutProductAdded(&product)

	return &core.CreateProductResult{
		Id:   int64(product.ID),
		Code: product.Code,
	}, nil
}

func (s productServer) EditProduct(ctx context.Context, req *core.EditProductRequest) (*core.EditProductReply, error) {
	updates := models.Product{}.Decode(req.Product)
	if updates.ID == 0 {
		return &core.EditProductReply{Error: "zero product id"}, nil
	}

	product, err := models.GetProductByID(uint64(updates.ID), "InstagramImages", "Items")
	if err != nil {
		return &core.EditProductReply{Error: err.Error()}, nil
	}

	if req.EditorId != 0 {
		models.IsUserSupplierOrSeller(req.EditorId, uint64(product.ShopID))
		return &core.EditProductReply{Forbidden: true}, nil
	}

	err = db.New().Model(product).Update(updates).Error
	if err != nil {
		log.Errorf("failed to update product %v: %v", updates.ID, err)
		return &core.EditProductReply{Error: err.Error()}, nil
	}
	return &core.EditProductReply{}, nil
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

	product, err := models.GetProductByID(uint64(req.ProductId))

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
		go nats.Publish("core.product.flush", product.ID)
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

func (s productServer) GetLikedBy(ctx context.Context, in *core.GetLikedByRequest) (*core.GetLikedByReply, error) {
	var ids []uint64
	err := db.New().
		Select("DISTINCT product_id").
		Table("users_products").
		Where("user_id = ?", in.UserId).
		Where("deleted_at IS NULL").
		Pluck("product_id", &ids).Error
	if err != nil {
		return nil, errors.New("db error")
	}
	return &core.GetLikedByReply{ProductIds: ids}, nil
}

func (s productServer) GetLastProductID(ctx context.Context, in *core.GetLastProductIDRequest) (*core.GetLastProductIDReply, error) {

	var out []uint64

	err := db.New().
		Select("id").
		Table("products_product").
		Where("shop_id = ?", in.ShopId).
		Order("id desc").
		Limit(1).
		Pluck("id", &out).
		Error
	if err != nil {
		return nil, err
	}

	if len(out) != 1 {
		return &core.GetLastProductIDReply{}, nil
	}

	return &core.GetLastProductIDReply{
		Id: out[0],
	}, nil
}

func (s productServer) DelProduct(ctx context.Context, in *core.DelProductRequest) (*core.DelProductReply, error) {

	err := db.New().
		Where("id = ?", in.ProductId).
		Delete(&models.Product{}).
		Error
	if err != nil {
		return nil, err
	}

	go nats.Publish("core.product.flush", in.ProductId)

	return &core.DelProductReply{
		Success: true,
	}, nil
}
