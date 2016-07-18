package views

import (
	"errors"
	"fmt"
	"net/http"
	"api/core"
	"api/soso"
)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"create", "email", CreateEmail},
	)
}

// Parameters:
//
//   * email (required string)
//   * comment (optional string)
// Return:
//   201 if created
func CreateEmail(c *soso.Context) {
	req := c.RequestMap

	email, email_present := req["email"].(string)
	msg_type, type_present := req["type"].(string)
	comment, comment_present := req["comment"].(string)

	if !comment_present || !email_present || !type_present {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Required field empty"))
		return
	}
	var subject = fmt.Sprintf("Question type: %s", msg_type)
	if err := core.SendEmail(subject, comment, email); err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{})
}
