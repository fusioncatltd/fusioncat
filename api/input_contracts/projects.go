package input_contracts

type CreateModifyProjectApiInputContract struct {
	Name        string `json:"name" binding:"required,min=1,max=45,alphanum"`
	Description string `json:"description"`
}
