package activity

import (
	"fmt"
	"strings"

	"github.com/qor/admin"
)

func getPrimaryKey(context *admin.Context, record interface{}) string {
	db := context.GetDB()

	var primaryValues []string
	for _, field := range db.NewScope(record).PrimaryFields() {
		primaryValues = append(primaryValues, fmt.Sprint(field.Field.Interface()))
	}
	return strings.Join(primaryValues, "::")
}

// GetActivities get activities for selected types
func GetActivities(context *admin.Context, types ...string) ([]QorActivity, error) {
	var activities []QorActivity
	db := context.GetDB().Order("id asc").Where("resource_id = ? AND resource_type = ?", context.Resource.GetPrimaryValue(context.Request), context.Resource.ToParam())

	var inTypes, notInTypes []string
	for _, t := range types {
		if strings.HasPrefix(t, "-") {
			notInTypes = append(notInTypes, strings.TrimPrefix(t, "-"))
		} else {
			inTypes = append(inTypes, t)
		}
	}

	if len(inTypes) > 0 {
		db = db.Where("type IN (?)", inTypes)
	}

	if len(notInTypes) > 0 {
		db = db.Where("type NOT IN (?)", notInTypes)
	}

	err := db.Find(&activities).Error
	return activities, err
}

// CreateActivity creates an activity for this context
func CreateActivity(context *admin.Context, activity *QorActivity, result interface{}) error {
	var activityResource = context.Admin.GetResource("QorActivity")

	// fill in necessary activity fields
	activity.ResourceType = context.Resource.ToParam()
	activity.ResourceID = getPrimaryKey(context, result)
	activity.CreatorName = context.CurrentUser.DisplayName()

	return activityResource.CallSave(activity, context.Context)
}
