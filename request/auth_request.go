package request

type Register struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPassword struct {
	Email string `json:"email"`
}

type PostForgot struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type VerifyUser struct {
	Token string `json:"token"`
}
