package cerror

import "net/http"

const (
	//auth
	ErrPhoneNumberReceive    = "정확한 전화번호를 입력해주세요"
	ErrPasswordNotMatched    = "비밀번호가 일치하지 않습니다"
	ErrRefreshTokenInvalid   = "인증 정보가 만료되었습니다. 다시 로그인을 시도 해주세요"
	ErrNumPasswordFailExceed = "비밀번호 최대 실패 횟수를 초과했습니다."
	ErrMultiLogin            = "다른 기기에서 접속이 확인되었습니다."
)

func BadRequest() CustomError {
	return newCustomError(
		http.StatusBadRequest,
		http.StatusText(http.StatusBadRequest),
	)
}

// err status 400 메시지와 함께 반환
func BadRequestWithMsg(message string) CustomError {
	return newCustomError(http.StatusBadRequest, message)
}
