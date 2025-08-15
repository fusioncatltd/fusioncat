package input_contracts

type CreateAppApiInputContract struct {
	Name        string `json:"name" binding:"required,min=1,max=45,alphanum_with_underscore_and_dots"`
	Description string `json:"description"`
}