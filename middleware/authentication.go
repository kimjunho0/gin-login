package middleware

var userok = User{
	Username: "username",
	Password: "password",
}

type User struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
