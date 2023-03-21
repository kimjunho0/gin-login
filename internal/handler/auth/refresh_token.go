package auth

type NewRefreshToken struct {
	RefreshToken string `json:"refresh_token"binding:"required"`
}

func CreateRefresh() {

}
