// fork from https://github.com/tektoncd/cli/blob/v0.13.1/pkg/log/log.go

package log

const (
	LogTypePipeline = "pipeline"
	LogTypeTask     = "task"
)

// Log represents data to write on log channel
type Log struct {
	Pipeline string
	Task     string
	Step     string
	Log      string
}
