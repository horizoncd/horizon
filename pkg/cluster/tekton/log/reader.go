// fork from https://github.com/tektoncd/cli/blob/v0.13.1/pkg/log/reader.go

package log

import (
	"fmt"

	he "g.hz.netease.com/horizon/core/errors"
	perrors "g.hz.netease.com/horizon/pkg/errors"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/options"
	"github.com/tektoncd/cli/pkg/pods"
	"github.com/tektoncd/cli/pkg/pods/stream"
)

type Reader struct {
	run      string
	ns       string
	clients  *cli.Clients
	streamer stream.NewStreamerFunc
	allSteps bool
	tasks    []string
	steps    []string
	logType  string
	task     string
	number   int
}

func NewReader(logType string, opts *options.LogOptions) (*Reader, error) {
	streamer := pods.NewStream
	if opts.Streamer != nil {
		streamer = opts.Streamer
	}

	cs, err := opts.Params.Clients()
	if err != nil {
		return nil, perrors.Wrap(he.ErrParamInvalid, err.Error())
	}

	var run string
	switch logType {
	case LogTypePipeline:
		run = opts.PipelineRunName
	case LogTypeTask:
		run = opts.TaskrunName
	}

	return &Reader{
		run:      run,
		ns:       opts.Params.Namespace(),
		clients:  cs,
		streamer: streamer,
		allSteps: opts.AllSteps,
		tasks:    opts.Tasks,
		steps:    opts.Steps,
		logType:  logType,
	}, nil
}

func (r *Reader) Read() (<-chan Log, <-chan error, error) {
	switch r.logType {
	case LogTypePipeline:
		return r.readPipelineLog()
	case LogTypeTask:
		return r.readTaskLog()
	}
	return nil, nil, fmt.Errorf("unknown log type")
}

func (r *Reader) setNumber(number int) {
	r.number = number
}

func (r *Reader) setRun(run string) {
	r.run = run
}

func (r *Reader) setTask(task string) {
	r.task = task
}

func (r *Reader) clone() *Reader {
	c := *r
	return &c
}
