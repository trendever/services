package models

import (
	"github.com/jinzhu/gorm"
	"proto/core"
	"time"
)

//Decode fill filter from core.SearchProductRequest
func (f *ProductFilter) Decode(req *core.SearchProductRequest) error {
	f.Keyword = req.Keyword
	f.Tags = req.Tags
	f.Limit = int(req.Limit)
	f.Offset = int(req.GetOffset())
	f.FromID = req.GetFromId()
	f.IsSaleOnly = req.IsSaleOnly
	f.Direction = req.OffsetDirection
	f.UserID = req.UserId
	f.ShopID = req.ShopId

	//We don't want execute  queries for users or shops which are not exist
	switch {
	case req.InstagramName != "":
		//Shop has priority
		id, err := FindShopIDByInstagramName(req.InstagramName)
		if err != nil {
			id, err := FindUserIDByInstagramName(req.InstagramName)
			if err != nil {
				return err
			}
			f.UserID = uint64(id)
		} else {
			f.ShopID = uint64(id)
		}
	}

	return nil
}

//Encode converts Product to core.Product
func (p *Product) Encode() *core.Product {
	// converts core product model to protoProduct
	ret := &core.Product{
		Id:    int64(p.ID),
		Title: p.Title,
		Code:  p.Code,

		Supplier:   p.Shop.Encode(),
		SupplierId: int64(p.ShopID),

		Mentioned:   p.MentionedBy.PublicEncode(),
		MentionedId: int64(p.MentionedByID),
		Items:       ProductItems(p.Items).Encode(),

		InstagramImageCaption:   p.InstagramImageCaption,
		InstagramImageUrl:       p.InstagramImageURL,
		InstagramImageId:        p.InstagramImageID,
		InstagramImageHeight:    uint32(p.InstagramImageHeight),
		InstagramImageWidth:     uint32(p.InstagramImageWidth),
		InstagramLink:           p.InstagramLink,
		InstagramPublishedAtAgo: int64(time.Since(p.InstagramPublishedAt).Seconds()),
		InstagramLikesCount:     int32(p.InstagramLikesCount),
		InstagramImages:         ImageCandidates(p.InstagramImages).Encode(),
		ChatMessage:             p.ChatMessage,
		WebShopUrl:              p.WebShopURL,

		IsSale: p.IsSale,

		LikedBy: Users(p.LikedBy).PublicEncode(),
	}
	if !p.InstagramPublishedAt.IsZero() {
		ret.InstagramPublishedAt = p.InstagramPublishedAt.Unix()
	}
	return ret
}

//Decode converts core.Product to Product
func (p Product) Decode(cp *core.Product) Product {
	ret := Product{
		Model: gorm.Model{
			ID: uint(cp.Id),
		},

		Title: cp.Title,
		Code:  cp.Code,

		Shop:   Shop{}.Decode(cp.Supplier),
		ShopID: uint(cp.SupplierId),

		MentionedBy:   User{}.Decode(cp.Mentioned),
		MentionedByID: uint(cp.MentionedId),

		Items: ProductItems{}.Decode(cp.Items),

		InstagramImageCaption: cp.InstagramImageCaption,
		InstagramImageURL:     cp.InstagramImageUrl,
		InstagramImageID:      cp.InstagramImageId,
		InstagramImageHeight:  uint(cp.InstagramImageHeight),
		InstagramImageWidth:   uint(cp.InstagramImageWidth),
		InstagramLink:         cp.InstagramLink,
		InstagramLikesCount:   int(cp.InstagramLikesCount),
		InstagramImages:       ImageCandidates{}.Decode(cp.InstagramImages),
		ChatMessage:           cp.ChatMessage,
		WebShopURL:            cp.WebShopUrl,

		IsSale: cp.IsSale,
	}
	if cp.InstagramPublishedAt != 0 {
		ret.InstagramPublishedAt = time.Unix(cp.InstagramPublishedAt, 0)
	}
	return ret
}

//Encode converts ProductItem to core.ProductItem
func (i ProductItem) Encode() *core.ProductItem {
	return &core.ProductItem{
		Id:            int64(i.ID),
		Name:          i.Name,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Tags:          Tags(i.Tags).Encode(),
	}
}

//Decode converts core.ProductItem to ProductItem
func (i ProductItem) Decode(ic *core.ProductItem) ProductItem {
	return ProductItem{
		Model: gorm.Model{
			ID: uint(ic.Id),
		},
		Name:          ic.Name,
		Price:         ic.Price,
		DiscountPrice: ic.DiscountPrice,
		Tags:          Tags{}.Decode(ic.Tags),
	}
}

//Encode converts to []*core.Product
func (p Products) Encode() []*core.Product {
	results := make([]*core.Product, len(p))
	for i, v := range p {
		results[i] = v.Encode()
	}
	return results
}

//Encode converts to []*core.ProductItem
func (pi ProductItems) Encode() []*core.ProductItem {
	results := make([]*core.ProductItem, len(pi))
	for i, v := range pi {
		results[i] = v.Encode()
	}
	return results
}

//Decode converts to ProductItems
func (pi ProductItems) Decode(ii []*core.ProductItem) ProductItems {
	pi = make(ProductItems, len(ii))
	decoder := ProductItem{}
	for i, v := range ii {
		pi[i] = decoder.Decode(v)
	}
	return pi
}

//Encode converts to []*core.ImageCandidate
func (ic ImageCandidates) Encode() []*core.ImageCandidate {
	results := make([]*core.ImageCandidate, len(ic))
	for i, v := range ic {
		results[i] = v.Encode()
	}
	return results
}

//Decode converts to ImageCandidates
func (ic ImageCandidates) Decode(ii []*core.ImageCandidate) ImageCandidates {
	ic = make(ImageCandidates, len(ii))
	decoder := ImageCandidate{}
	for i, v := range ii {
		ic[i] = decoder.Decode(v)
	}
	return ic
}

//Encode converts ImageCandidate to core.ImageCandidate
func (i ImageCandidate) Encode() *core.ImageCandidate {
	return &core.ImageCandidate{
		Id:   int64(i.ID),
		Url:  i.URL,
		Name: i.Name,
	}
}

//Decode converts core.ImageCandidate to ImageCandidate
func (i ImageCandidate) Decode(ic *core.ImageCandidate) ImageCandidate {
	return ImageCandidate{
		ID:   uint(ic.Id),
		URL:  ic.Url,
		Name: ic.Name,
	}
}
