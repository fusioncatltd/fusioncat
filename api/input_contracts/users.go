package input_contracts

type SignInSignUpApiInputContract struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
