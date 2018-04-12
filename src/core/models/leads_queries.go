package models

import (
	"common/db"
	"common/log"
	"core/api"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"proto/chat"
	"proto/core"
	"reflect"
	"strings"
	"time"
	"utils/rpc"
)

// FindLeadByID returns Lead with id
func FindLeadByID(id uint) (Lead, error) {
	var lead Lead
	err := db.
		New().
		Where("id = ?", id).
		Find(&lead).Error

	return lead, err
}

//GetUserLeads returns user leads
// If shopID in not zero, limit query to this shop,
// limit == 0 -> 20,
// fromUpdatedAt is optional.
// Results are sorted by updated_at, ascending if direction is true, descending otherwise
func GetUserLeads(user *User, roles []core.LeadUserRole, shopID uint64, limit uint64, fromUpdatedAt int64, direction bool) (leads LeadCollection, err error) {
	//First, we must find leads with passed parameters
	scope := db.New().
		Table("products_leads as pl").
		Where("pl.deleted_at IS NULL").
		Select("pl.id")

	if fromUpdatedAt != 0 {
		t := time.Unix(0, fromUpdatedAt)
		if direction {
			scope = scope.Where("pl.chat_updated_at > ?", t)
		} else {
			scope = scope.Where("pl.chat_updated_at < ?", t)
		}
	}

	// we don't want show leads for sellers before first customer message
	ignoreForSeller := []string{
		core.LeadStatus_EMPTY.String(),
		core.LeadStatus_NEW.String(),
	}

	var relatedSellerShops []uint64
	var relatedSupplierShops []uint64

	if user.SuperSeller {
		if len(roles) == 1 && roles[0] == core.LeadUserRole_CUSTOMER {
			scope = scope.Where("pl.customer_id = ?", user.ID)
		} else {
			scope = scope.Where("pl.customer_id = ? OR state NOT IN (?)", user.ID, ignoreForSeller)
		}
	} else { // !user.SuperSeller
		var or []string
		var orArgs []interface{}

		if hasLeadRole(core.LeadUserRole_CUSTOMER, roles) {
			or = append(or, "pl.customer_id = ?")
			orArgs = append(orArgs, user.ID)
		}

		if hasLeadRole(core.LeadUserRole_SELLER, roles) {
			relatedSellerShops, err = GetShopsIDWhereUserIsSeller(user.ID)
			if err != nil {
				return nil, err
			}
			if len(relatedSellerShops) > 0 {
				or = append(or, "(state NOT IN (?) AND pl.shop_id IN (?))")
				orArgs = append(orArgs, ignoreForSeller, relatedSellerShops)
			}
		}

		if hasLeadRole(core.LeadUserRole_SUPPLIER, roles) {
			relatedSupplierShops, err = GetShopsIDWhereUserIsSupplier(user.ID)
			if err != nil {
				return nil, err
			}
			if len(relatedSupplierShops) > 0 {
				or = append(or, "(state NOT IN (?) AND pl.shop_id IN (?))")
				orArgs = append(orArgs, ignoreForSeller, relatedSupplierShops)
			}
		}

		switch len(or) {
		case 0:
			return nil, nil
		case 1:
			scope = scope.Where(or[0], orArgs...)
		default:
			scope = scope.Where("("+strings.Join(or, ") OR (")+")", orArgs...)
		}
	}

	if limit != 0 {
		scope = scope.Limit(int(limit))
	} else {
		scope = scope.Limit(20)
	}

	if shopID != 0 {
		scope = scope.Where("pl.shop_id = ?", shopID)
	}

	if direction {
		scope = scope.Order("pl.chat_updated_at asc")
	} else {
		scope = scope.Order("pl.chat_updated_at desc")
	}

	var ids []uint64
	err = scope.Pluck("pl.id", &ids).Error
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	// @REFACTOR Part with products still looks doubtful
	var rows *sql.Rows
	//Second, we must search related products through products_leads_items
	rows, err = db.New().
		Table("products_product_item as ppi").
		Select("DISTINCT ppi.product_id, pli.lead_id").
		Joins("LEFT JOIN  products_leads_items pli ON ppi.id = pli.product_item_id").
		Where("pli.lead_id in (?)", ids).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productIDs := []uint{}
	productMap := map[uint][]uint{}
	leadMap := map[uint][]*Product{}
	for rows.Next() {
		var id, leadID uint
		err = rows.Scan(&id, &leadID)
		if err != nil {
			return nil, err
		}
		productIDs = append(productIDs, id)
		productMap[id] = append(productMap[id], leadID)
	}
	products := []*Product{}
	err = db.New().Preload("Items").Preload("InstagramImages").Where("id in (?)", productIDs).Find(&products).Error
	if err != nil {
		return nil, err
	}
	for _, product := range products {
		lIDs, ok := productMap[product.ID]
		if !ok {
			continue
		}
		for _, lID := range lIDs {
			leadProducts, ok := leadMap[lID]
			if !ok {
				leadProducts = []*Product{}
			}
			leadProducts = append(leadProducts, product)
			leadMap[lID] = leadProducts
		}

	}
	leadsScope := db.New().Where("id in (?)", ids).Preload("Shop").Preload("Shop.Sellers").Preload("Shop.Supplier").Preload("Customer")
	if direction {
		leadsScope = leadsScope.Order("chat_updated_at asc")
	} else {
		leadsScope = leadsScope.Order("chat_updated_at desc")
	}
	err = leadsScope.Find(&leads).Error

	if err != nil {
		return nil, err
	}
	//Third, merge all
	for _, lead := range leads {
		leadProducts, ok := leadMap[lead.ID]
		if ok {
			lead.Products = leadProducts
		}

		lead.UserRole = lead.RoleOf(user)
	}

	return leads, nil
}

//GetUserLead returns users's lead by lead ID
func GetUserLead(user *User, leadID uint64) (*Lead, error) {
	lead, err := GetLead(leadID, 0, "Customer", "Shop", "Shop.Supplier", "Shop.Sellers")
	if err != nil {
		return nil, err
	}

	lead.UserRole = lead.RoleOf(user)
	if lead.UserRole == core.LeadUserRole_UNKNOWN {
		return nil, errors.New("forbidden")
	}

	lead.Products, err = GetLeadProducts(lead)
	if err != nil {
		return nil, err
	}

	return lead, nil
}

// FindActiveLead searches active lead for shop and customer,
// returns nil result for shops with SeparateLeads if productID isn't presented in active leads
func FindActiveLead(shopID, customerID, productID uint64) (*Lead, error) {
	lead := &Lead{}
	scope := db.New().
		Model(&Lead{}).
		Preload("Customer").
		Preload("Shop").
		Preload("Shop.Supplier").
		Joins("JOIN products_shops shop ON shop.id = products_leads.shop_id").
		Where("shop.id = ?", shopID).
		Where("products_leads.customer_id = ?", customerID).
		Where("products_leads.state IN (?)", []string{
			core.LeadStatus_EMPTY.String(),
			core.LeadStatus_NEW.String(),
			core.LeadStatus_IN_PROGRESS.String(),
			core.LeadStatus_SUBMITTED.String(),
		}).
		Where(`
		NOT shop.separate_leads
		OR EXISTS (
			SELECT 1 FROM products_leads_items related
			JOIN products_product_item item
			ON item.id = related.product_item_id AND item.deleted_at IS NULL
			WHERE related.lead_id = products_leads.id
			AND item.product_id = ?
		)`, productID).
		First(lead)
	if scope.RecordNotFound() {
		return nil, nil
	}

	if scope.Error != nil {
		return nil, scope.Error
	}

	return lead, nil
}

//CreateLead creates new lead
func CreateLead(protoLead *core.Lead, shopID uint) (lead *Lead, err error) {
	lead = Lead{}.Decode(protoLead)

	err = db.New().First(&lead.Customer, lead.CustomerID).Error
	if err != nil {
		return nil, err
	}

	lead.ShopID = shopID
	err = db.New().First(&lead.Shop, shopID).Error
	if err != nil {
		return nil, err
	}
	err = db.New().First(&lead.Shop.Supplier, lead.Shop.SupplierID).Error
	if err != nil {
		return nil, err
	}

	var members = []*chat.Member{
		{
			UserId:      uint64(lead.CustomerID),
			Role:        chat.MemberRole_CUSTOMER,
			Name:        lead.Customer.GetName(),
			InstagramId: lead.Customer.InstagramID,
		},
	}

	if sellers, err := GetSellersByShopID(lead.ShopID); err == nil {
		for _, seller := range sellers {
			members = append(members, &chat.Member{
				UserId: uint64(seller.ID),
				Role:   chat.MemberRole_SELLER,
				Name:   seller.GetName(),
			})
		}
	}

	lead.ConversationID, err = CreateChat(
		members,
		genChatCaption(lead),
		protoLead.DirectThread,
		lead.Shop.Supplier.InstagramID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %v", err)
	}

	lead.State = core.LeadStatus_NEW.String()
	if err := db.New().Create(&lead).Error; err != nil {
		//here we have a chat leak... this is rare case anyway
		return nil, err
	}
	return lead, nil
}

func CreateChat(members []*chat.Member, caption, directThread string, primaryInstagram uint64) (chatID uint64, err error) {
	context, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := api.ChatServiceClient.CreateChat(context, &chat.NewChatRequest{
		Chat: &chat.Chat{
			Members:      members,
			Caption:      caption,
			DirectThread: directThread,
		},
		PrimaryInstagram: primaryInstagram,
	})
	switch {
	case err != nil:
		return 0, err

	case resp.Error != nil:
		return 0, errors.New(resp.Error.Message)

	default:
		return resp.Chat.Id, nil
	}
}

func genChatCaption(lead *Lead) string {
	template := &OtherTemplate{}
	ret := db.New().Find(template, "template_id = ?", "chat_caption")
	if ret.RecordNotFound() {
		return ""
	}
	if ret.Error != nil {
		log.Errorf("failed to load template: %v", ret.Error)
		return ""
	}
	result, err := template.Execute(map[string]interface{}{
		"lead": lead,
	})
	if err != nil {
		log.Errorf("failed to execute template: %v", err)
		return ""
	}
	text, ok := result.(string)
	if !ok {
		log.Errorf("expected string, but got " + reflect.TypeOf(text).Name())
		return ""
	}
	return text
}

//AppendLeadItems adds new items to the lead, and returns count of new items
func AppendLeadItems(lead *Lead, items []ProductItem) (int, error) {
	oldCount := db.New().Model(lead).Association("ProductItems").Count()
	err := db.New().Model(lead).Association("ProductItems").Append(items).Error
	newCount := db.New().Model(lead).Association("ProductItems").Count()
	return newCount - oldCount, err
}

//GetLead returns lead by id or conversation_id
func GetLead(id, conversationID uint64, preloads ...string) (*Lead, error) {
	searchLead := &Lead{
		Model: gorm.Model{
			ID: uint(id),
		},
		ConversationID: conversationID,
	}

	query := db.New().Where(searchLead)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	query.Find(searchLead)
	return searchLead, query.Error
}

//GetFullLeadByID returns lead by id with all preloads
func GetFullLeadByID(id uint64) (*Lead, error) {
	return GetLead(id, 0, "Customer", "Shop", "Shop.Supplier", "Shop.Sellers", "ProductItems")
}

//TouchLead updates only chat_updated_at field, without call any callbacks
func TouchLead(conversationID uint64) error {
	return db.New().Model(&Lead{}).Where("conversation_id = ?", conversationID).UpdateColumn("chat_updated_at", time.Now()).Error
}

//GetLeadProducts returns products for the lead
func GetLeadProducts(lead *Lead) ([]*Product, error) {
	rows, err := db.New().
		Table("products_product_item as ppi").
		Select("DISTINCT ppi.product_id").
		Joins("LEFT JOIN  products_leads_items pli ON ppi.id = pli.product_item_id").
		Where("pli.lead_id = ?", lead.ID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	productIDs := []uint{}
	for rows.Next() {
		var id uint
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		productIDs = append(productIDs, id)
	}

	products := []*Product{}
	err = db.New().
		Preload("Items").
		Preload("InstagramImages").
		Where("id in (?)", productIDs).
		Find(&products).
		Error
	if err != nil {
		return nil, err
	}

	return products, nil
}

// GetUsersForLead returns every userID that can possibly join this chat
func GetUsersForLead(lead *Lead) ([]uint64, error) {

	var users = map[uint]bool{}

	shop, err := GetShopByID(lead.ShopID)
	if err != nil {
		return nil, err
	}

	users[lead.CustomerID] = true
	users[shop.SupplierID] = true

	sellers, err := GetSellersByShopID(lead.ShopID)
	if err != nil {
		return nil, err
	}

	for _, seller := range sellers {
		users[seller.ID] = true
	}

	superSellers, err := GetSuperSellersIDs()
	if err != nil {
		return nil, err
	}

	for _, seller := range superSellers {
		users[seller] = true
	}

	var out = make([]uint64, 0, len(users))
	for k := range users {
		out = append(out, uint64(k))
	}

	return out, nil

}
