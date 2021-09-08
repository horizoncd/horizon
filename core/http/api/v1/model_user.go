package v1

// User 登录用户信息
type User struct {
	Name string `json:"name,omitempty"`
	Email string `json:"email"`
}
