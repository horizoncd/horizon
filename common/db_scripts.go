package common

/* sql about gitlab */
const (
	GitlabQuery       = "select * from gitlab"
	GitlabQueryByName = "select * from gitlab where name = ?"
)

/* sql about template */
const (
	TemplateQuery                             = "select * from template"
	TemplateReleaseQueryByTemplateName        = "select * from template_release where template_name = ?"
	TemplateReleaseQueryByTemplateNameAndName = "select * from template_release where template_name = ? and name = ?"
)

/* sql about user */
const (
	// UserQueryByOIDC ...
	UserQueryByOIDC = "select * from user where oidc_id = ? and oidc_type = ?"
)
