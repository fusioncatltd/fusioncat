package input_contracts

type CreateSchemaApiInputContract struct {
	Name        string `json:"name" binding:"required,min=1,max=45,alphanum_with_underscore"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required,oneof=jsonschema"`
	Schema      string `json:"schema" binding:"required,valid_json_schema"`
}

type ModifySchemaApiInputContract struct {
	Schema string `json:"schema" binding:"required,valid_json_schema"`
}
