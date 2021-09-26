package common

/* sql about gitlab */
const (
	GitlabQuery       = "select * from gitlab where deleted_at is null"
	GitlabQueryByName = "select * from gitlab where name = ? and deleted_at is null"
)

/* sql about template */
const (
	TemplateQuery                      = "select * from template where deleted_at is null"
	TemplateReleaseQueryByTemplateName = "select * from template_release " +
		"where template_name = ? and deleted_at is null"
	TemplateReleaseQueryByTemplateNameAndName = "select * from template_release " +
		"where template_name = ? and name = ? and deleted_at is null"
)

/* sql about user */
const (
	// UserQueryByOIDC ...
	UserQueryByOIDC = "select * from user where oidc_id = ? and oidc_type = ?"
)

const (
	ApplicationQueryByName = "select * from application where name = ? and deleted_at is null"
)
