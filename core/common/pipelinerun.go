package common

const (
	PipelineQueryByStatus = "status"

	MessageQueryBySystem = "system"

	MessagePipelinerunStopped   = "stopped pipelinerun"
	MessagePipelinerunExecuted  = "executed pipelinerun"
	MessagePipelinerunCancelled = "cancelled pipelinerun"
	MessagePipelinerunReady     = "marked pipelinerun as ready to execute"
)
