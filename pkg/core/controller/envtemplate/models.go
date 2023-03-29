package envtemplate

type EnvTemplate struct {
	Application map[string]interface{} `json:"application"`
	Pipeline    map[string]interface{} `json:"pipeline"`
}

type UpdateEnvTemplateRequest struct {
	*EnvTemplate
}

type GetEnvTemplateResponse struct {
	*EnvTemplate
}
