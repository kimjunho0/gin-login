package cerror

import "fmt"

type CustomError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func newCustomError(statusCode int, errorMessage string) CustomError {
	return CustomError{
		StatusCode: statusCode,
		Message:    errorMessage,
	}
}

func (error *CustomError) Error() string {
	return fmt.Sprintf("[%d]%s", error.StatusCode, error.Message)
}

type CustomError400 struct {
	StatusCode int    `json:"status_code" example:"400"`
	Message    string `json:"message" example:"입력하신 부분을 다시 확인해주세요"`
}
type CustomError401 struct {
	StatusCode int    `json:"status_code" example:"401"`
	Message    string `json:"message" example:"인증에 실패했습니다 다시 로그인 해주세요"`
}
type CustomError403 struct {
	StatusCode int    `json:"status_code" example:"403"`
	Message    string `json:"message" example:"권한이 없습니다"`
}
type CustomError500 struct {
	StatusCode int    `json:"status_code" example:"500"`
	Message    string `json:"message" example:"예기치 않은 오류"`
}
