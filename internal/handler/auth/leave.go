package auth

import (
	"fmt"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/models"
	"gin-login/pkg/cerror"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"runtime/debug"
)

// @tags auth
// @Summary delete_user
// @Description 회원 탈퇴
// @Accept json
// @Produce json
// @Param auth-token header string true "access token"
// @Param pwd path string true "패스워드"
// @Success 200
// @Failure 400 {object} cerror.CustomError400
// @Failure 401 {object} cerror.CustomError401
// @Failure 500 {object} cerror.CustomError500
// @Router /api/auth/leave/{pwd} [DELETE]
func Leave(c *gin.Context) {

	defer func() {
		if err := recover(); err != nil {
			log.Printf(fmt.Sprintf("%v \n %v", err, string(debug.Stack())))
		}
		if c.Writer.Written() {
			return
		}
		c.JSON(http.StatusBadRequest, cerror.CustomError{
			StatusCode: 500,
			Message:    "Unexpected internal server error!",
		})
	}()

	//path 에서 password 받아오기
	pwd, isExist := c.Params.Get("pwd")
	if !isExist {
		panic(cerror.BadRequest())
	}
	//userid 받아오기
	userId := middleware.GetReqManagerIdFromToken(c.Request)

	//transaction 시작
	tx := migrate.DB.Begin()
	defer tx.Rollback()

	//입력한 패스워드, 토큰에 맞는 아이디
	var user = models.User{
		Id:       userId,
		Password: pwd,
	}
	//기존 pw 가져오기
	var pw string

	if err := tx.Model(&models.User{}).
		Where("id = ?", user.Id).
		Select("password").
		Take(&pw).Error; err != nil {
		panic(cerror.DBErr(err))
	}
	//password 비교
	if !PasswordCompare(pw, user.Password) {
		c.JSON(http.StatusBadRequest, "비번 틀림")
		panic(cerror.BadRequestWithMsg("비번 틀림"))
	}

	if err := tx.Delete(&user).Error; err != nil {
		panic(cerror.DBErr(err))
	}

	tx.Commit()
	//transaction 끝
	c.JSON(http.StatusOK, "유저삭제 완료")
}
