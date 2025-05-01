package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func MatchUserTypeToUid(c *gin.Context, userID string) (err error) {
	userType := c.GetString("user_type")
	uID := c.GetString("user_id")
	err = nil
	if userType == "USER" && uID != userID {
		err = errors.New("unauthorized access to this resource")
		return err
	}
	err = CheckUserType(c, userType)
	return err

}

func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil
	if userType != role {
		err = errors.New("unauthorized access to this resource")
		return err
	}
	return err
}
