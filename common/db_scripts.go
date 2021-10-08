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
	UserSearch      = "select * from user where name like ? or full_name like ? or email like ? limit ? offset ?"
	UserSearchCount = "select count(1) from user where name like ? or full_name like ? or email like ?"
)

/* sql about member */
const (
	MemberUpdate = "select * from member where id = ? and deleted_at is null"
)
