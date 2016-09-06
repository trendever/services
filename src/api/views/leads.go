package views

import (
	"api/soso"
	"errors"
	"net/http"

	"api/api"
	"api/chat"
	"api/models"
	"fmt"
	pchat "proto/chat"
	"proto/core"
	"strings"
	"utils/log"
	"utils/rpc"
)

var leadServiceClient = core.NewLeadServiceClient(api.CoreConn)

func init() {

	SocketRoutes = append(SocketRoutes,
		soso.Route{"create", "lead", CreateLead},
		soso.Route{"list", "lead", GetUserLeads},
		soso.Route{"retrieve", "lead", GetUserLead},
		soso.Route{"event", "lead", SetLeadStatus},
		soso.Route{"get_cancel_reasons", "lead", GetCancelReasons},
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

func GetUserLeads(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap
	var limit uint64
	var from_updated_at int64
	var roles []core.LeadUserRole

	if value, ok := req["roles"].(string); ok {
		sroles := strings.Split(value, ",")
		for _, r := range sroles {
			rid, ok := core.LeadUserRole_value[strings.ToUpper(r)]
			if ok {
				roles = append(roles, core.LeadUserRole(rid))
			}
		}
	}

	if value, ok := req["from_updated_at"].(float64); ok {
		from_updated_at = int64(value)
	}

	if value, ok := req["limit"].(float64); ok {
		limit = uint64(value)

	}

	if len(roles) == 0 {
		groups := map[string][]core.LeadUserRole{
			"customer": {core.LeadUserRole_CUSTOMER},
			"seller":   {core.LeadUserRole_SELLER, core.LeadUserRole_SUPPLIER},
		}

		results := map[string]*models.Leads{}

		for name, roles := range groups {
			var err error
			results[name], err = getUserLeads(c.Token.UID, roles, limit, false, from_updated_at)
			if err != nil {
				c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
				return
			}
		}

		c.SuccessResponse(results)
		return

	}
	direction := false
	if value, ok := req["direction"].(bool); ok {
		direction = value
	}

	leads, err := getUserLeads(c.Token.UID, roles, limit, direction, from_updated_at)
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
		log.Error(fmt.Errorf("failed to get cancel reasons: %v", err))
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
			log.Error(fmt.Errorf("Chat with id %v not found", lead.ConversationId))
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

	var messages *pchat.ChatHistoryReply
	if lead.LeadInfo.ConversationId != 0 {
		lead.Chat, err = getChat(lead.LeadInfo.ConversationId, c.Token.UID)
		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}
		messages, err = chat.GetMessages(lead.LeadInfo.ConversationId, c.Token.UID, 0, 12)
		if err != nil {
			c.ErrorResponse(http.StatusNotFound, soso.LevelError, err)
			return
		}
	}

	c.SuccessResponse(map[string]interface{}{
		"lead":     lead,
		"messages": messages.Messages,
		"error":    messages.Error,
	})
}

func getUserLeads(uid uint64, roles []core.LeadUserRole, limit uint64, direction bool, from_updated_at int64) (*models.Leads, error) {
	request := &core.UserLeadsRequest{
		UserId:        uid,
		Role:          roles,
		Limit:         limit,
		FromUpdatedAt: from_updated_at,
		Direction:     direction,
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
