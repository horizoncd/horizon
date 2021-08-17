package login

// User 登录用户信息
type User struct {
	HzNumber    string `json:"hzNumber"` // TODO(gjq) set this field a real hzNumber
	ID          int    `json:"id"`
	Login       bool   `json:"login"` // TODO(gjq) delete this field
	MailAddress string `json:"mailAddress,omitempty"`
	Name        string `json:"name,omitempty"`
	NickName    string `json:"nickName,omitempty"`
	SuperAdmin  bool   `json:"superAdmin"`
}
