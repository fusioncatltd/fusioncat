package input_contracts

type CreateServerApiInputContract struct {
	Name        string `json:"name" binding:"required,min=1,max=45,alphanum_with_underscore_and_dots"`
	Description string `json:"description"`
	Protocol    string `json:"protocol" binding:"required,async_protocol"`
}

type CreateResourceApiInputContract struct {
	Name         string `json:"name" binding:"required,min=1,max=100,resource_uri"`
	Mode         string `json:"mode" binding:"required,oneof=read write readwrite"`
	ResourceType string `json:"resource_type" binding:"required,oneof=topic exchange queue table endpoint"`
	Description  string `json:"description"`
}

type CreateResourceBindApiInputContract struct {
	SourceResourceID string `json:"source_resource_id" binding:"required,uuid"`
	TargetResourceID string `json:"target_resource_id" binding:"required,uuid"`
}