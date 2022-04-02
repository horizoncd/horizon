package context

// nolint
type contextKey struct{ payload uint8 }

var ContextQueryOnCondition = &contextKey{}
var ContextEmails = &contextKey{}
var ContextDirectMemberOnly = &contextKey{}
