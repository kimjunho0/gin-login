package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type needlogin struct {
	phoneNumber string `json:"phone_number" binding:"required"`
	password    string `json:"password" binding:"required"`
}

func login(c *gin.Context) {
	var login needlogin
	if err := c.ShouldBindJSON(&login); err != nil {
		fmt.Sprintf("login binding error", err)
	}

}

func PasswordCompare() {

}
