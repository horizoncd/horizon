package slo

type PipelineSLO struct {
	Name                string  `json:"name"`                // 英文名，build、deploy等
	DisplayName         string  `json:"displayName"`         // 中文名，构建、发布等
	Count               int     `json:"count"`               // 次数
	RequestAvailability float64 `json:"requestAvailability"` // 请求可用率 99.9945
	RequestSlo          float64 `json:"requestSlo"`          // 请求可用率SLO 99.99
	RTAvailability      float64 `json:"rtAvailability"`      // RT可用率 99.9825
	RTSlo               float64 `json:"rtSlo"`               // RT可用率SLO 99.98
	RT                  uint    `json:"rt"`                  // RT临界值 60 单位秒
}
