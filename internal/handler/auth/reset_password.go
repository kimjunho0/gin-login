package auth

import (
	"gin-login/internal/constants"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ResetModel struct {
	NewPassword string `json:"new_password" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
}
type IfSuccessReset struct {
	Message string           `json:"message"`
	Status  constants.Status `json:"status"`
}

const (
	success = "비밀번호 리셋 성공"
	failed  = "비밀번호 리셋 실패"
)

// @tags auth
// @Summary 패스워드 초기화
// @Description 비밀번호 초기화
// @Accept json
// @Produce json
// @Param num path string true "전화번호"
// @Param body body auth.ResetModel true "바꿀 비밀번호, 현재 비밀번호"
// @Success 200 {object} auth.IfSuccessReset
// @Failure 400
// @Router /api/auth/reset-password/{num} [PATCH]
func ResetPassword(c *gin.Context) {
	var body ResetModel
	if err := c.ShouldBind(&body); err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
	}
	phoneNumber, isExist := c.Params.Get("num")
	if !isExist {
		panic(cerror.BadRequest())
	}

	PasswordValidity(body.NewPassword, phoneNumber)
	Pw := models.User{
		Password:     PasswordHash(body.NewPassword),
		PhoneNumber:  phoneNumber,
		RefreshToken: RefreshToken(),
	}

	//fail 구조체에 성공유무 넣기
	fail := IfSuccessReset{
		Message: failed,
		Status:  constants.StatusFail,
	}
	// Todo : 예전 비밀번호와 폰번호 일치하는지 확인 후에 새로운 비밀번호로 변경 --완료--

	//transaction start
	tx := migrate.DB.Begin()
	defer tx.Rollback()

	//전화번호로 password 가져옴
	user := middleware.TakeManagerInformation(phoneNumber, "password")
	//폰번호의 비번과 입력한 비번이 일치하는지 확인
	if !PasswordCompare(user.Password, body.OldPassword) {

		//fail 구조체에 성공유무 넣기
		fail := IfSuccessReset{
			Message: failed,
			Status:  constants.StatusFail,
		}
		c.JSON(http.StatusBadRequest, fail)
		panic(cerror.BadRequestWithMsg("비밀번호 틀림"))
	}

	//phone number 에 맞는 password, refresh token, password fail 초기화 혹은 값을 바꿈
	if err := tx.Model(&Pw).
		Where("phone_number = ?", Pw.PhoneNumber).
		Updates(map[string]interface{}{
			"password":          Pw.Password,
			"refresh_token":     Pw.RefreshToken,
			"num_password_fail": 0,
		}).Error; err != nil {

		c.JSON(http.StatusBadRequest, fail)

		panic(cerror.DBErr(err))

	}

	tx.Commit()
	//transaction 끝

	success := IfSuccessReset{
		Message: success,
		Status:  constants.StatusOk,
	}

	c.JSON(http.StatusOK, success)
}
