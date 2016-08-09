package views

import (
	"api/api"
	"api/soso"
	"errors"
	"io"
	"net/http"
	"proto/core"
	"strconv"
	"strings"
	"utils/rpc"
)

var pushTokensServiceClient = core.NewPushTokensServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"add", "push_tokens", AddPushToken},
		soso.Route{"del", "push_tokens", DelPushToken},
		soso.Route{"get", "push_tokens", GetPushTokens},
	)
	http.HandleFunc("/addtoken", AddTokenHTTP)
}

func AddPushToken(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	token, ok := req["token"].(string)
	if !ok {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Empty token"))
		return
	}
	typeString, _ := req["type"].(string)
	typeId, ok := typeIdFromString(typeString)
	if !ok {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Unknown token type"))
		return
	}

	about, _ := req["about"].(string)

	request := &core.AddTokenRequest{Token: &core.TokenInfo{
		UserId: c.Token.UID,
		Token:  token,
		Type:   typeId,
		About:  about,
	}}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := pushTokensServiceClient.AddToken(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.Error))
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"status": "success",
	})
}

func DelPushToken(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	id, ok := c.RequestMap["token_id"].(float64)
	if !ok {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Empty id"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := pushTokensServiceClient.DelToken(ctx, &core.DelTokenRequest{
		UserId:  c.Token.UID,
		TokenId: uint64(id),
	})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.Error))
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"status": "success",
	})
}

func GetPushTokens(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := pushTokensServiceClient.GetTokens(ctx, &core.GetTokensRequest{
		UserId: c.Token.UID,
	})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.Error))
		return
	}

	type jsonToken struct {
		Id    uint64
		Type  string
		Token string
		About string
	}
	tokens := []jsonToken{}
	for _, t := range res.Tokens {
		tokens = append(tokens, jsonToken{
			Id:    t.Id,
			Type:  strings.ToLower(core.TokenType_name[int32(t.Type)]),
			Token: t.Token,
			About: t.About,
		})
	}
	c.SuccessResponse(map[string]interface{}{
		"tokens": tokens,
	})
}

func AddTokenHTTP(w http.ResponseWriter, r *http.Request) {
	user_id, err := strconv.ParseUint(r.FormValue("user_id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "user_id is missing or invalid")
		return
	}
	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "token is missing")
		return
	}
	typeId, ok := typeIdFromString(r.FormValue("type"))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "type is missing or unknown")
		return
	}
	about := r.FormValue("about")

	request := &core.AddTokenRequest{Token: &core.TokenInfo{
		UserId: user_id,
		Token:  token,
		Type:   typeId,
		About:  about,
	}}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := pushTokensServiceClient.AddToken(ctx, request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}
	if res.Error != "" {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, res.Error)
		return
	}

	io.WriteString(w, "OK")
}

func typeIdFromString(str string) (core.TokenType, bool) {
	str = strings.ToLower(str)
	var typeId core.TokenType
	knownType := false
	for key, name := range core.TokenType_name {
		if strings.ToLower(name) == str {
			typeId = core.TokenType(key)
			knownType = true
		}
	}
	return typeId, knownType
}
