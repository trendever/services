package utils

import (
	"errors"
	"proto/auth"
	"utils/rpc"
	"core/api"
)

// GetTokenData returns the decoded token
// @TODO: move it to utils?
func GetTokenData(token string) (*auth.Token, error) {
	request := &auth.TokenDataRequest{Token: token}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := api.AuthServiceClient.GetTokenData(ctx, request)

	if err != nil {
		return nil, err
	}
	if resp.ErrorCode != auth.ErrorCodes_NO_ERRORS {
		return nil, errors.New("Invalid or expired token")
	}
	return resp.Token, nil
}
