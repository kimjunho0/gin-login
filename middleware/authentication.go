package middleware

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}
type AccessAndRefreshResponse struct {
}
