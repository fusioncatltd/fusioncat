package input_contracts

type CreateMessageApiInputContract struct {
	Name          string `json:"name" binding:"required,min=1,max=45,alphanum_with_underscore"`
	Description   string `json:"description"`
	SchemaID      string `json:"schema_id" binding:"required,uuid"`
	SchemaVersion int    `json:"schema_version" binding:"required,min=1"`
}