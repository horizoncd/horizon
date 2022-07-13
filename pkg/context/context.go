package context

// nolint
type contextKey struct{ payload uint8 }

var MemberQueryOnCondition = &contextKey{}
var MemberEmails = &contextKey{}
var MemberDirectMemberOnly = &contextKey{}

var TemplateWithRelease = &contextKey{}
var TemplateOnlyRefCount = &contextKey{}

var ReleaseSyncToRepo = &contextKey{}
