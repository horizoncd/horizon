package context

// nolint
type contextKey struct{ payload uint8 }

var (
	MemberQueryOnCondition = &contextKey{} // the context key for member query conditions
	MemberEmails           = &contextKey{} // the context key for member emails
	MemberRole             = &contextKey{} // the context key for member roles
	MemberDirectMemberOnly = &contextKey{} // the context key for member only
)

var (
	TemplateWithRelease     = &contextKey{} // the context key for template with release
	TemplateOnlyRefCount    = &contextKey{} // the context key for template only ref count
	TemplateWithFullPath    = &contextKey{} // the context key for template with full path
	TemplateListSelfOnly    = &contextKey{} // the context key for listing templates self only
	TemplateListRecursively = &contextKey{} // the context key for listing templates recursively
)

var (
	ReleaseSyncToRepo = &contextKey{} // the context key for release sync to repo
	JWTTokenString    = &contextKey{} // the context key for JWT token string
)
