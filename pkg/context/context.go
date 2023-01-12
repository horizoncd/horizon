package context

// nolint
type contextKey struct{ payload uint8 }

var MemberQueryOnCondition = &contextKey{}
var MemberEmails = &contextKey{}
var MemberRole = &contextKey{}
var MemberDirectMemberOnly = &contextKey{}

var TemplateWithRelease = &contextKey{}
var TemplateOnlyRefCount = &contextKey{}
var TemplateWithFullPath = &contextKey{}
var TemplateListSelfOnly = &contextKey{}
var TemplateListRecursively = &contextKey{}

var ReleaseSyncToRepo = &contextKey{}

var JWTTokenString = &contextKey{}
