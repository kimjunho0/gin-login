package login

import (
	"gin-login/internal/constants"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"gin-login/redis/session"
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
	fail    = "비밀번호 리셋 실패"
)

// @tags auth
// @Summary reset-password
// @Description 비밀번호 초기화
// @Accept json
// @Produce json
// @Param num path string true "전화번호"
// @Param body body login.ResetModel true "바꿀 비밀번호, 현재 비밀번호"
// @Success 200 {object} login.IfSuccessReset
// @Failure 400 {object} cerror.CustomError400
// @Failure 401 {object} cerror.CustomError401
// @Failure 500 {object} cerror.CustomError500
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
	if IfPhoneNumberIncludeChar(phoneNumber) {
		panic(cerror.BadRequestWithMsg(errPhoneNumberPasswordEqual))
	}
	Pw := models.User{
		Password:     PasswordHash(body.NewPassword),
		PhoneNumber:  phoneNumber,
		RefreshToken: RefreshToken(),
	}

	// Todo : 예전 비밀번호와 폰번호 일치하는지 확인 후에 새로운 비밀번호로 변경

	//transaction start
	tx := migrate.DB.Begin()
	defer tx.Rollback()

	resp := IfSuccessReset{
		Status:  constants.StatusOk,
		Message: success,
	}

	//전화번호로 password 가져옴
	user := middleware.TakeManagerInformation(phoneNumber, "password")
	//폰번호의 비번과 입력한 비번이 일치하는지 확인
	if !PasswordCompare(user.Password, body.OldPassword) {
		panic(cerror.BadRequestWithMsg("비밀번호가 틀렸습니다."))
	}
	//바꿀 비번이 규칙에 맞는지
	PasswordValidity(body.NewPassword, phoneNumber)

	//phone number 에 맞는 password, refresh token, password fail 초기화 혹은 값을 바꿈
	if result := tx.Model(&Pw).
		Where("phone_number = ?", Pw.PhoneNumber).
		Updates(map[string]interface{}{
			"password":          Pw.Password,
			"refresh_token":     Pw.RefreshToken,
			"num_password_fail": 0,
		}); result.Error != nil {
		panic(cerror.DBErr(result.Error))
		//업데이트 된 레코드 수가 1개가 아니면 = 1개 이상이거나 없으면 실패 메시지 반환
	} else if result.RowsAffected != 1 {
		resp.Status = constants.StatusFail
		resp.Message = fail
	}

	tx.Commit()
	//transaction 끝

	//logout from all device
	if err := migrate.DB.
		Select([]string{"id"}).
		Where("phone_number = ?", user.PhoneNumber).
		Take(&user).Error; err != nil {
		panic(cerror.BadRequestWithMsg(err.Error()))
	}
	session.Logout(user.Id)

	c.JSON(http.StatusOK, resp)
}
