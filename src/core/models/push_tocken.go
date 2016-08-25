package models

import (
	"github.com/jinzhu/gorm"

	"core/db"
	"errors"
	"fmt"
	proto "proto/core"
)

type PushToken struct {
	gorm.Model
	UserId uint            `gorm:"index;unique_index:compose_unique"`
	Type   proto.TokenType `gorm:"unique_index:compose_unique"`
	Token  string          `gorm:"type:text;not null;unique_index:compose_unique"`
	About  string          `gorm:"type:text"`
}

func (t PushToken) Validate(db *gorm.DB) {
	if t.UserId == 0 {
		db.AddError(errors.New("user should not be empty"))
	}
	if _, ok := proto.TokenType_name[int32(t.Type)]; !ok {
		db.AddError(fmt.Errorf("unknown token type %v", t.Type))
	}
	if t.Token == "" {
		db.AddError(errors.New("token itself should not be empty"))
	}
}

// mockable interface to db
type PushTokensRepository interface {
	AddToken(token *PushToken) error
	DelToken(id, user_id uint) error
	GetTokens(user_id uint) ([]PushToken, error)
	InvalidateTokens(t proto.TokenType, tokens []string)
	UpdateToken(t proto.TokenType, oldToken, newToken string)
}

type pushTokensRepositoryImpl struct{}

func GetPushTokensRepository() PushTokensRepository {
	return &pushTokensRepositoryImpl{}
}

func (*pushTokensRepositoryImpl) AddToken(token *PushToken) error {
	if token.ID != 0 {
		return errors.New("id should be empty in add request")
	}
	res := db.New().Save(&token)
	if res.Error != nil {
		return fmt.Errorf("failed to add token: %v", res.Error)
	}
	return nil
}

func (*pushTokensRepositoryImpl) DelToken(id, userId uint) error {
	var token PushToken
	res := db.New().First(&token, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("failed to load token: %v", res.Error)
	}
	if token.UserId != userId {
		return fmt.Errorf("user %v isn't owner of token %v", userId, id)
	}
	res = db.New().Delete(&token)
	if res.Error != nil {
		return fmt.Errorf("failed to delete token: %v", res.Error)
	}
	return nil
}

func (*pushTokensRepositoryImpl) InvalidateTokens(t proto.TokenType, tokens []string) {
	if tokens == nil || len(tokens) == 0 {
		return
	}
	db.New().Where("type = ?", t).Where("token IN (?)", tokens).Delete(PushToken{})
}

func (*pushTokensRepositoryImpl) UpdateToken(t proto.TokenType, oldToken, newToken string) {
	db.New().Model(&PushToken{}).Where("type = ?", t).Where("token = ?", oldToken).Update("token", newToken)
}

func (*pushTokensRepositoryImpl) GetTokens(userId uint) ([]PushToken, error) {
	var tokens []PushToken
	res := db.New().Find(&tokens, "user_id = ?", userId)
	if res.Error != nil && !res.RecordNotFound() {
		return nil, fmt.Errorf("failed to load tokens: %v", res.Error)
	}
	return tokens, nil
}

/*
 * RPC-related stuff
 */

func (t *PushToken) Encode() *proto.TokenInfo {
	return &proto.TokenInfo{
		Id:     uint64(t.ID),
		UserId: uint64(t.UserId),
		Type:   t.Type,
		Token:  t.Token,
		About:  t.About,
	}
}

func (t PushToken) Decode(tp *proto.TokenInfo) *PushToken {
	return &PushToken{
		Model:  gorm.Model{ID: uint(tp.Id)},
		UserId: uint(tp.UserId),
		Type:   tp.Type,
		Token:  tp.Token,
		About:  tp.About,
	}
}
