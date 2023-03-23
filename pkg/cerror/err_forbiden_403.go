package cerror

import "net/http"

//헤더의 auth-token 부분의 토큰값이 요청 되었으나 이 access 토큰이 권한이 없는경우(로그인 시간 초과 or 토큰이 잘못됨)

func Forbidden() CustomError {
	return newCustomError(
		http.StatusForbidden,
		http.StatusText(http.StatusForbidden),
	)
}

func ForbiddenWithMsg(message string) CustomError {
	return newCustomError(http.StatusForbidden, message)
}
