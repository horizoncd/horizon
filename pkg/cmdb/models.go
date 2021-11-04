package cmdb

type PriorityType int

const (
	P0 PriorityType = 0
	P1 PriorityType = 1
	P2 PriorityType = 2
)

type Account struct {
	Account     string `json:"account"`
	AccountType string `json:"type"` // user or group
}

type CreateApplicationRequest struct {
	Name     string       `json:"name"`
	ParentID int          `json:"parentId"`
	Priority PriorityType `json:"priority"`
	Admin    []Account    `json:"admin"`
}

type DeleteApplication struct {
	appName string
}

type Env string

const (
	Dev    Env = "dev"
	Test   Env = "test"
	Reg    Env = "reg"
	Pre    Env = "pre"
	beta   Env = "beta"
	Perf   Env = "perf"
	Online Env = "online"
)

type ClusterStatusType string

const (
	StatusReady  ClusterStatusType = "ready"
	StatusOnline ClusterStatusType = "Online"
)

type CreateClusterRequest struct {
	Name                string            `json:"name"`
	ApplicationName     string            `json:"applicationName"`
	Env                 Env               `json:"env"`
	ClusterServerStatus ClusterStatusType `json:"clusterServerStatus"`
	ClusterStyle        string            `json:"clusterStyle"`
	Admin               []Account         `json:"admin"`
}

type CommonResp struct {
	RT   interface{} `json:"rt,omitempty"`
	Code int         `json:"code"`
}
