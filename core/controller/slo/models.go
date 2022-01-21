package slo

type PipelineSLO struct {
	Name                string // 英文名，build、deploy等
	DisplayName         string // 中文名
	Count               int    // 次数
	RequestAvailability int    // 请求可用率
	RTAvailability      int    // RT可用率
	RT                  uint   // RT临界值
}
