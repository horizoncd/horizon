package slo

type PipelineSLO struct {
	Name                string  // 英文名，build、deploy等
	DisplayName         string  // 中文名，构建、发布等
	Count               int     // 次数
	RequestAvailability float64 // 请求可用率 99.9945
	RTAvailability      float64 // RT可用率 99.9925
	RT                  uint    // RT临界值 60 单位秒
}
