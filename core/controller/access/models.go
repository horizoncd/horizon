package access

type API struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}

type ReviewResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

// ReviewRequest provide apis for access review.
type ReviewRequest struct {
	APIs []API `json:"apis"`
}
