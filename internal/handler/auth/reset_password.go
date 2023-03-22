package auth

import (
	"gin-login/migrate"
	"gin-login/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ResetModel struct {
	Password    string `json:"password" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}
type IfSuccessReset struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func ResetPassword(c *gin.Context) {
	var body ResetModel
	if err := c.ShouldBind(&body); err != nil {
		panic("reset binding error")
	}
	Pw := models.User{
		Password:     PasswordHash(body.Password),
		PhoneNumber:  body.PhoneNumber,
		RefreshToken: RefreshToken(),
	}

	//transaction start
	tx := migrate.DB.Begin()
	defer tx.Rollback()
	if err := migrate.DB.Model(&Pw).
		Where("phone_number = ?", Pw.PhoneNumber).
		Updates(map[string]interface{}{
			"password":          Pw.Password,
			"refresh_token":     Pw.RefreshToken,
			"num_password_fail": 0,
		}).Error; err != nil {
		panic("reset password transaction error")
		c.JSON(http.StatusBadRequest, IfSuccessReset{
			Message: "Fail Reset",
			Status:  "Not Registered",
		})
	}
	tx.Commit()
	//transaction 끝
	c.JSON(http.StatusOK, IfSuccessReset{
		Message: "Success",
		Status:  "Success",
	})

}