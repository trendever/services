package views

import (
	"api/api"
	"api/chat"
	"api/models"
	"common/log"
	"common/soso"
	"errors"
	"net/http"
	"proto/core"
	"strings"
	"utils/rpc"
)

var leadServiceClient = core.NewLeadServiceClient(api.CoreConn)

func init() {

	SocketRoutes = append(SocketRoutes,
		soso.Route{"lead", "create", CreateLead},
		soso.Route{"lead", "list", GetUserLeads},
		soso.Route{"lead", "retrieve", GetUserLead},
		soso.Route{"lead", "event", SetLeadStatus},
		soso.Route{"lead", "get_cancel_reasons", GetCancelReasons},
	)
}

// CreateLead parameters:
//
//   * id (required int) product id
//   * auth parameters:
//   * instagram_username (optional string)
//   * phone number (optional string)
// At least one of (instagram_username or phone) must be set; or user must be logged in
// Return:
//   lead_id, ID of the created opportunity
func CreateLead(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	req := c.RequestMap

	// Step #1: convert input from interface{} and check it

	id, idPresent := req["id"].(float64)
	action_f, _ := req["action"].(float64)
	action := core.LeadAction(action_f)
	if _, ok := core.LeadAction_name[int32(action)]; !ok {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("unknown action"))
		return
	}

	var (
		userID int64
		err    error
	)

	userID = int64(c.Token.UID)

	if !idPresent {
		err = errors.New("Product ID not set")
	}

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	// Context is responsible for timeouts
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	leadRes, err := leadServiceClient.CreateLead(ctx, &core.Lead{
		Source:     "website",
		CustomerId: userID,
		ProductId:  int64(id),
		Action:     action,
	})

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	lead := &models.Lead{LeadInfo: *leadRes.Lead}

	lead.Chat, err = getChat(lead.LeadInfo.ConversationId, c.Token.UID)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"lead": lead,
	})
}

func GetUserLeads(c *soso.Context, arg *struct {
	Roles         string   `json:"roles"`
	RelatedShop   uint64   `json:"shop_id"`
	FromUpdatedAt int64    `json:"from_updated_at"`
	Limit         uint64   `json:"limit"`
	Direction     bool     `json:"direction"`
	Tags          []uint64 `json:"tags"`
}) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	var roles []core.LeadUserRole

	for _, r := range strings.Split(arg.Roles, ",") {
		rid, ok := core.LeadUserRole_value[strings.ToUpper(r)]
		if ok {
			roles = append(roles, core.LeadUserRole(rid))
		}
	}

	if len(roles) == 0 {
		groups := map[string][]core.LeadUserRole{
			"customer": {core.LeadUserRole_CUSTOMER},
			"seller":   {core.LeadUserRole_SELLER, core.LeadUserRole_SUPPLIER},
		}

		results := map[string]*models.Leads{}

		for name, roles := range groups {
			var err error
			results[name], err = getUserLeads(
				c.Token.UID, roles, arg.RelatedShop, arg.Tags,
				arg.Limit, arg.Direction, arg.FromUpdatedAt)
			if err != nil {
				c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
				return
			}
		}

		c.SuccessResponse(results)
		return
	}

	leads, err := getUserLeads(
		c.Token.UID, roles, arg.RelatedShop, arg.Tags,
		arg.Limit, arg.Direction, arg.FromUpdatedAt)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	//
	c.SuccessResponse(map[string]interface{}{
		"leads": leads,
	})
}

func GetCancelReasons(c *soso.Context) {
	// @TODO cache it?
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := leadServiceClient.GetCancelReasons(ctx, &core.GetCancelReasonsRequest{})
	if err != nil {
		log.Errorf("failed to get cancel reasons: %v", err)
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"reasons": resp.Reasons,
	})
}

func SetLeadStatus(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap
	request := &core.SetLeadStatusRequest{
		UserId: c.Token.UID,
	}

	if value, ok := req["lead_id"].(float64); ok {
		request.LeadId = uint64(value)
	} else {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("lead_id is required"))
		return
	}

	if value, ok := req["event"].(string); ok {
		event, ok := core.LeadStatusEvent_value[value]
		if !ok {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Unknown lead event"))
			return
		}
		request.Event = core.LeadStatusEvent(event)
	} else {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Event is required"))
		return
	}

	if value, ok := req["cancel_reason"].(float64); ok {
		request.CancelReason = uint64(value)
	}
	if value, ok := req["status_commet"].(string); ok {
		request.StatusComment = value
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := leadServiceClient.SetLeadStatus(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}
	//send new status for related users
	lead := &models.Lead{LeadInfo: *resp.Lead}
	if lead.LeadInfo.ConversationId != 0 {
		lead.Chat, err = getChat(lead.LeadInfo.ConversationId, c.Token.UID)

		if err != nil {
			log.Error(err)
		}

		if lead.Chat != nil {
			r := map[string]interface{}{
				"lead": lead,
			}
			remote_ctx := soso.NewRemoteContext("lead", "retrieve", r)

			go chat.BroadcastMessage(lead.Chat.Members, c, remote_ctx)
		} else {
			log.Errorf("Chat with id %v not found", lead.ConversationId)
		}

	}
	//
	c.SuccessResponse(map[string]interface{}{
		"lead": lead,
	})
}

func GetUserLead(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	req := c.RequestMap
	request := &core.GetLeadRequest{
		UserId: c.Token.UID,
	}

	if value, ok := req["lead_id"].(float64); ok {
		request.SearchBy = &core.GetLeadRequest_Id{Id: uint64(value)}
	}

	if value, ok := req["conversation_id"].(float64); ok {
		request.SearchBy = &core.GetLeadRequest_ConversationId{ConversationId: uint64(value)}
	}

	if request.GetConversationId() == 0 && request.GetId() == 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("one of lead_id or conversation_id is required"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := leadServiceClient.GetLead(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	if resp.Lead == nil {
		c.ErrorResponse(http.StatusNotFound, soso.LevelError, errors.New("Lead not found"))
		return
	}

	lead := &models.Lead{
		LeadInfo: *resp.Lead,
	}

	ret := map[string]interface{}{
		"lead": lead,
	}
	if lead.LeadInfo.ConversationId != 0 {
		lead.Chat, err = getChat(lead.LeadInfo.ConversationId, c.Token.UID)
		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}
		messages, err := chat.GetMessages(lead.LeadInfo.ConversationId, c.Token.UID, 0, 12)
		if err != nil {
			c.ErrorResponse(http.StatusNotFound, soso.LevelError, err)
			return
		}
		ret["messages"] = messages.Messages
		ret["error"] = messages.Error
	}

	c.SuccessResponse(ret)
}

func getUserLeads(uid uint64, roles []core.LeadUserRole, relatedShop uint64, tags []uint64, limit uint64, direction bool, fromUpdatedAt int64) (*models.Leads, error) {
	request := &core.UserLeadsRequest{
		UserId:        uid,
		Role:          roles,
		Limit:         limit,
		FromUpdatedAt: fromUpdatedAt,
		Direction:     direction,
		RelatedShop:   relatedShop,
		Tags:          tags,
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := leadServiceClient.GetUserLeads(ctx, request)
	if err != nil {

		return nil, err
	}
	chIDs := []uint64{}
	for _, l := range resp.Leads {
		if l.ConversationId != 0 {
			chIDs = append(chIDs, l.ConversationId)
		}
	}

	leads := &models.Leads{}
	if chats, err := getChats(chIDs, uid); err == nil {
		leads.Fill(resp.Leads, chats)
	} else {
		return nil, err
	}

	return leads, nil
}

func getLeadInfo(userID, leadID uint64) (*core.LeadInfo, error) {

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := leadServiceClient.GetLead(ctx, &core.GetLeadRequest{
		UserId:   userID,
		SearchBy: &core.GetLeadRequest_Id{Id: leadID},
	})
	if err != nil {
		return nil, err
	}

	// now checks
	if resp.Lead == nil {
		return nil, errors.New("Lead not found")
	}

	return resp.Lead, nil

}

func getUserRole(userID, leadID, chatID uint64) (core.LeadUserRole, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := leadServiceClient.GetUserRole(ctx, &core.GetUserRoleRequest{
		UserId:         userID,
		LeadId:         leadID,
		ConversationId: chatID,
	})
	switch {
	case err != nil:
		return core.LeadUserRole_UNKNOWN, err
	case resp.Error != "":
		return core.LeadUserRole_UNKNOWN, errors.New(resp.Error)
	default:
		return resp.Role, nil
	}
}
