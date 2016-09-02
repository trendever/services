package views

import (
	"core/api"
	"core/models"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"utils/db"
	"utils/log"
)

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		core.RegisterTagServiceServer(s, tagService{})
	})
}

type tagService struct{}

func (ts tagService) GetMainTags(ctx context.Context, req *core.GetMainTagsRequest) (*core.TagSearchResult, error) {
	var tags models.Tags

	err := db. // order is not necessary, qor/sorting.Sorting will make it for us
			New().
			Where("main = ?", true).
			Where("hidden = ?", false).
			Where("EXISTS (SELECT 1 FROM products_product_item_tags relation WHERE relation.tag_id = products_tag.id)").
			Limit(int(req.Limit)).
			Find(&tags).
			Error

	if err != nil {
		return nil, err
	}

	result := tags.Encode()

	return &core.TagSearchResult{Result: result}, nil
}

// Returns related tags
// For a search with tags req.Tags
// we want to retrieve a related tag list.
// Related tag is each tag that is specified in found products,
// but not included in the search query.
func (ts tagService) GetRelatedTags(ctx context.Context, req *core.GetRelatedTagsRequest) (*core.TagSearchResult, error) {

	tags, err := models.RelatedTags(req.Tags, int(req.Limit))

	if err != nil {
		log.Error(err)
		return nil, err
	}

	result := models.Tags(tags).Encode()

	return &core.TagSearchResult{Result: result}, nil
}
