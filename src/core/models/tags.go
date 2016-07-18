package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/qor/sorting"
	"github.com/qor/validations"
	"proto/core"
)

// Tag model
type Tag struct {
	gorm.Model
	sorting.Sorting

	GroupID sql.NullInt64
	Group   TagGroup
	Main    bool
	Hidden  bool   `gorm:"not null" sql:"default:false"`
	Name    string `gorm:"unique"`
}

//Tags is a collection of Tags
type Tags []Tag

// TagGroup is a container for multiple tags
type TagGroup struct {
	gorm.Model
	Name string `gorm:"unique"`
}

// TableName for tags
func (t Tag) TableName() string {
	return "products_tag"
}

// ResourceName for tags
func (t Tag) ResourceName() string {
	return "Tags"
}

// TableName for groups
func (t TagGroup) TableName() string {
	return "products_tag_group"
}

// ResourceName for tags
func (t TagGroup) ResourceName() string {
	return "TagGroup"
}

// Validate tag
func (t Tag) Validate(db *gorm.DB) {
	if t.Name == "" {
		db.AddError(validations.NewError(t, "Name", "Name can not be empty"))
	}
}

// Validate group
func (t TagGroup) Validate(db *gorm.DB) {
	if t.Name == "" {
		db.AddError(validations.NewError(t, "Name", "Name can not be empty"))
	}
}

//Encode converts to core.Tag
func (t Tag) Encode() *core.Tag {
	return &core.Tag{
		Id:   int64(t.ID),
		Name: t.Name,
	}
}

//Decode converts to Tag
func (t Tag) Decode(tag *core.Tag) Tag {
	return Tag{
		Model: gorm.Model{
			ID: uint(tag.Id),
		},
		Name: tag.Name,
	}
}

//Encode converts collection to core.Tag collection
func (t Tags) Encode() []*core.Tag {
	results := make([]*core.Tag, len(t))
	for i, v := range t {
		results[i] = v.Encode()
	}
	return results
}

//Decode converts core.Tag collection to Tags collection
func (t Tags) Decode(tags []*core.Tag) Tags {
	t = make(Tags, len(tags))
	decoder := Tag{}
	for i, v := range tags {
		t[i] = decoder.Decode(v)
	}
	return t
}
