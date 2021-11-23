package cmdb

import "errors"

type AccountType string

const (
	User AccountType = "user"
)

type ClusterStyle string

const (
	Docker ClusterStyle = "docker"
)

type PriorityType int

const (
	P0 PriorityType = 0
	P1 PriorityType = 1
	P2 PriorityType = 2
)

type Account struct {
	Account     string      `json:"account"`
	AccountType AccountType `json:"type"` // user or group
}

type CreateApplicationRequest struct {
	Name     string       `json:"name"`
	ParentID int          `json:"parentId"`
	Priority PriorityType `json:"priority"`
	Admin    []Account    `json:"admin"`
}

type Env string

const (
	Dev       Env = "dev"
	Test      Env = "test"
	Reg       Env = "reg"
	Pre       Env = "pre"
	Beta      Env = "beta"
	Perf      Env = "perf"
	Online    Env = "online"
	UnKownEnv Env = "unknown"
)

func ToCmdbEnv(evn string) (Env, error) {
	switch Env(evn) {
	case Dev:
		return Dev, nil
	case Test:
		return Test, nil
	case Reg:
		return Reg, nil
	case Pre:
		return Pre, nil
	case Beta:
		return Beta, nil
	case Perf:
		return Perf, nil
	case Online:
		return Online, nil
	default:
		return UnKownEnv, errors.New("")
	}
}

type ClusterStatusType string

const (
	StatusReady  ClusterStatusType = "ready"
	StatusOnline ClusterStatusType = "online"
)
const (
	AutoAddContainer int = 1
)

type CreateClusterRequest struct {
	Name                string            `json:"name"`
	ApplicationName     string            `json:"applicationName"`
	Env                 Env               `json:"env"`
	ClusterServerStatus ClusterStatusType `json:"clusterServerStatus"`
	ClusterStyle        ClusterStyle      `json:"clusterStyle"`
	AutoAddDocker       int               `json:"autoAddDocker"`
	Admin               []Account         `json:"admin"`
}

type CommonResp struct {
	RT   interface{} `json:"rt,omitempty"`
	Code int         `json:"code"`
}
