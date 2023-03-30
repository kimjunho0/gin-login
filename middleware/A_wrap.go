package middleware

import (
	"bytes"
	"gin-login/internal/constants"
	"gin-login/pkg/cerror"
	"gin-login/tools"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

func WrapMiddleware(c *gin.Context) {
	if !strings.Contains(c.Request.URL.Path, "swagger") && !strings.Contains(c.Request.URL.Path, "static") {
		c.Writer.Header().Set("Content-Type", "application/json")
	}
	c.Writer.Header().Set(constants.HeaderCacheControl, constants.CacheNoStore)

	bodyByte, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyByte))

	defer func() {

		//error handling

		//r 로 panic 이 일어나서 recover 를 일으켰는지 확인하고
		//panic 이 일어났다면 아래 switch 문을 실행
		if r := recover(); r != nil {
			var customError cerror.CustomError
			switch r.(type) {
			case error:
				customError = cerror.CustomError{
					StatusCode: http.StatusInternalServerError,
					Message:    http.StatusText(http.StatusInternalServerError),
				}
			case string:
				customError = cerror.CustomError{
					StatusCode: http.StatusInternalServerError,
					Message:    http.StatusText(http.StatusInternalServerError),
				}
			case cerror.CustomError:
				customError = r.(cerror.CustomError)
			default:
				customError = cerror.CustomError{
					StatusCode: http.StatusInternalServerError,
					Message:    http.StatusText(http.StatusInternalServerError),
				}
			}
			c.JSON(customError.StatusCode, customError)
			tools.LogError(&customError)
			errorStack := string(debug.Stack())
			log.Println(errorStack)
			log.Println(customError)

			c.Abort()

		}
	}()

	c.Next()
}
