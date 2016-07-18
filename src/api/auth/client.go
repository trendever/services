package auth

import (
	"errors"
	auth_protocol "proto/auth"
	"utils/rpc"
	"api/api"
)

// Client is auth service client
var Client = auth_protocol.NewAuthServiceClient(api.AuthConn)

//GetTokenData returns the decoded token
func GetTokenData(token string) (*auth_protocol.Token, error) {
	request := &auth_protocol.TokenDataRequest{Token: token}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := Client.GetTokenData(ctx, request)

	if err != nil {
		return nil, err
	}
	if resp.ErrorCode != auth_protocol.ErrorCodes_NO_ERRORS {
		return nil, errors.New("Invalid or expired token")
	}
	return resp.Token, nil
}
