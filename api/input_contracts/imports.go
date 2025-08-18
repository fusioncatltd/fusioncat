package input_contracts

type ImportFileInputContract struct {
	YAML string `json:"yaml" binding:"required"`
}