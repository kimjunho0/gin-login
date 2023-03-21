package handler

type ReqRegister struct {
	PhoneNumber string `json:"phone_number" binding:"required,len=11"`
	Password    string `json:"password" binding:"required"`
	Name        string `json:"name" binding:"required"`
}

func register() {

}
