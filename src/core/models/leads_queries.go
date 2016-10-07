package models

import (
	"database/sql"
	"errors"
	"github.com/jinzhu/gorm"
	"proto/core"
	"strings"
	"time"
	"utils/db"
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
// @REFACTOR Code of this func looks non-nice. May be we should rewrite it
func GetUserLeads(user *User, roles []core.LeadUserRole, leadID uint64, limit uint64, fromUpdatedAt int64, direction bool) (leads LeadCollection, err error) {
	var relatedSellerShops []uint64
	var relatedSupplierShops []uint64
	if hasLeadRole(core.LeadUserRole_SELLER, roles) && !user.SuperSeller {
		relatedSellerShops, err = GetShopsIDWhereUserIsSeller(user.ID)
		if err != nil {
			return
		}
	}

	if hasLeadRole(core.LeadUserRole_SUPPLIER, roles) && !user.SuperSeller {
		relatedSupplierShops, err = GetShopsIDWhereUserIsSupplier(user.ID)
		if err != nil {
			return
		}
	}
	//First, we must find leads with passed parameters
	userID := user.ID

	scope := db.New().
		Table("products_leads as pl").
		Where("pl.deleted_at IS NULL").
		Select("pl.id, pl.customer_id, pl.shop_id")

	or := []string{}
	orArgs := []interface{}{}

	// we don't want show leads for sellers before first customer message
	ignoreForSeller := []string{
		core.LeadStatus_EMPTY.String(),
		core.LeadStatus_NEW.String(),
	}

	//we don't want to filter leads for super seller, he can see all
	if hasLeadRole(core.LeadUserRole_CUSTOMER, roles) && !user.SuperSeller {
		or = append(or, "pl.customer_id = ?")
		orArgs = append(orArgs, userID)
	}
	if hasLeadRole(core.LeadUserRole_SUPPLIER, roles) && !user.SuperSeller && len(relatedSupplierShops) > 0 {
		or = append(or, "(state NOT IN (?) AND pl.shop_id IN (?))")
		orArgs = append(orArgs, ignoreForSeller, relatedSupplierShops)
	}
	if hasLeadRole(core.LeadUserRole_SELLER, roles) && !user.SuperSeller && len(relatedSellerShops) > 0 {
		or = append(or, "(state NOT IN (?) AND pl.shop_id IN (?))")
		orArgs = append(orArgs, ignoreForSeller, relatedSellerShops)
	}

	//this mean we want get leads where super seller is customer
	if user.SuperSeller && len(roles) == 1 && hasLeadRole(core.LeadUserRole_CUSTOMER, roles) {
		or = append(or, "pl.customer_id = ?")
		orArgs = append(orArgs, userID)
	}

	switch {
	case len(or) == 0:
		//User does request for leads without a customer role, and he hasn't linked shops, and he is not a super seller
		//therefore we can't show for him anything
		if !user.SuperSeller {
			return
		}
		// super seller
		scope = scope.Where("pl.customer_id = ? OR state NOT IN (?)", userID, ignoreForSeller)
	case len(or) > 1:
		scope = scope.Where("("+strings.Join(or, " OR ")+")", orArgs...)
	case len(or) == 1:
		scope = scope.Where(or[0], orArgs...)
	}

	if limit != 0 {
		scope = scope.Limit(int(limit))
	} else {
		scope = scope.Limit(20)
	}

	if fromUpdatedAt != 0 {
		t := time.Unix(0, fromUpdatedAt)
		if direction {
			scope = scope.Where("pl.chat_updated_at > ?", t.Format(time.RFC3339Nano))
		} else {
			scope = scope.Where("pl.chat_updated_at < ?", t.Format(time.RFC3339Nano))
		}
	}

	if leadID != 0 {
		scope = scope.Where("pl.id = ?", leadID)
	}

	if direction {
		scope = scope.Order("pl.chat_updated_at asc")
	} else {
		scope = scope.Order("pl.chat_updated_at desc")
	}

	var rows *sql.Rows
	rows, err = scope.Rows()
	if err != nil {
		return
	}
	defer rows.Close()
	var ids []uint64
	rolesMap := make(map[uint64]core.LeadUserRole)
	sellerShopMap := makeUintMap(relatedSellerShops)
	supplierShopMap := makeUintMap(relatedSupplierShops)
	for rows.Next() {
		var id, customerID, shopID uint64
		err = rows.Scan(&id, &customerID, &shopID)
		if err != nil {
			return
		}
		ids = append(ids, id)
		rolesMap[id] = toRole(customerID, shopID, sellerShopMap, supplierShopMap, user)
	}
	if len(ids) == 0 {
		return
	}

	//Second, we must search related products through products_leads_items
	rows, err = db.New().
		Table("products_product_item as ppi").
		Select("DISTINCT ppi.product_id, pli.lead_id").
		Joins("LEFT JOIN  products_leads_items pli ON ppi.id = pli.product_item_id").
		Where("pli.lead_id in (?)", ids).Rows()
	if err != nil {
		return
	}
	defer rows.Close()

	productIDs := []uint{}
	productMap := map[uint][]uint{}
	leadMap := map[uint][]*Product{}
	for rows.Next() {
		var id, leadID uint
		err = rows.Scan(&id, &leadID)
		if err != nil {
			return
		}
		productIDs = append(productIDs, id)
		productMap[id] = append(productMap[id], leadID)
	}
	products := []*Product{}
	err = db.New().Preload("Items").Preload("InstagramImages").Where("id in (?)", productIDs).Find(&products).Error
	if err != nil {
		return
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
		return
	}
	//Third, merge all
	for _, lead := range leads {
		leadProducts, ok := leadMap[lead.ID]
		if ok {
			lead.Products = leadProducts
		}

		role, ok := rolesMap[uint64(lead.ID)]
		if ok {
			lead.UserRole = role
		}
	}

	return
}

func toRole(customerID, shopID uint64, sellersSh map[uint64]int, suppliersSh map[uint64]int, user *User) core.LeadUserRole {
	if customerID == uint64(user.ID) {
		return core.LeadUserRole_CUSTOMER
	}
	if _, ok := suppliersSh[shopID]; ok {
		return core.LeadUserRole_SUPPLIER
	}
	if _, ok := sellersSh[shopID]; ok {
		return core.LeadUserRole_SELLER
	}
	if user.SuperSeller {
		return core.LeadUserRole_SUPER_SELLER
	}
	return core.LeadUserRole_UNKNOWN
}

//GetUserLead returns users's lead by lead ID
func GetUserLead(user *User, leadID uint64) (*Lead, error) {
	leads, err := GetUserLeads(user, []core.LeadUserRole{
		core.LeadUserRole_CUSTOMER,
		core.LeadUserRole_SUPPLIER,
		core.LeadUserRole_SELLER,
	}, leadID, 0, 0, false)
	if err != nil {
		return nil, err
	}
	if len(leads) != 1 {
		return nil, errors.New("Lead not found")
	}

	return leads[0], nil
}

// FindActiveLead searches active lead for shop and customer
// returns nil result for shops with SeparateLeads
func FindActiveLead(shopID, customerID uint64) (*Lead, error) {
	lead := &Lead{}
	scope := db.New().
		Model(&Lead{}).
		Preload("Customer").
		Joins("JOIN products_shops shop ON shop.id = products_leads.shop_id AND NOT shop.separate_leads").
		Where(
			"shop.id = ? AND products_leads.customer_id = ? AND products_leads.state IN (?)",
			shopID,
			customerID,
			[]string{
				core.LeadStatus_EMPTY.String(),
				core.LeadStatus_NEW.String(),
				core.LeadStatus_IN_PROGRESS.String(),
			}).
		Find(lead)
	if scope.RecordNotFound() {
		return nil, nil
	}

	if scope.Error != nil {
		return nil, scope.Error
	}

	return lead, nil
}

//CreateLead creates new lead
func CreateLead(protoLead *core.Lead, shopID uint) (*Lead, error) {
	lead := Lead{}.Decode(protoLead)
	lead.State = core.LeadStatus_EMPTY.String()
	lead.ShopID = shopID
	// If customer id is not correct, it should throw an error
	if err := db.New().Create(&lead).Error; err != nil {
		return nil, err
	}
	return lead, nil
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
