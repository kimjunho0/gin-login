package cerror

import (
	"fmt"
	"net/http"
)

const (
	UnknownErrTitle = "일시적인 문제가 발생하였습니다. 잠시후 다시 시도해주세요"
)

func DBErr(err error) CustomError {
	fmt.Println(err)
	return newCustomError(http.StatusInternalServerError, UnknownErrTitle)
}
func Marshal(err error) CustomError {
	fmt.Println(err)
	return newCustomError(http.StatusInternalServerError, UnknownErrTitle)
}
func Unknown(err error) CustomError {
	fmt.Println(err)
	return newCustomError(http.StatusInternalServerError, UnknownErrTitle)
}
func RedisErr(err error) CustomError {
	fmt.Println(err)
	return newCustomError(http.StatusInternalServerError, UnknownErrTitle)
}
