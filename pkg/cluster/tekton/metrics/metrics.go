package metrics

import (
	"g.hz.netease.com/horizon/pkg/server/global"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	_prHistogram   *prometheus.HistogramVec
	_trHistogram   *prometheus.HistogramVec
	_stepHistogram *prometheus.HistogramVec
)

const (
	_environment = "environment"
	_step        = "step"
	_pipeline    = "pipeline"
	_task        = "task"
	_result      = "result"
	_template    = "template"
	_namespace   = "horizon"
	_subsystem   = ""
)

func init() {
	buckets := []float64{
		0, 5, 10, 20, 30, 40, 50, 60, 90, 120, 150, 180, 240, 300,
	}
	_prHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    prometheus.BuildFQName(_namespace, _subsystem, "pipelinerun_duration_seconds"),
		Help:    "PipelineRun duration info",
		Buckets: buckets,
	}, []string{_environment, _template, _pipeline, _result})

	_trHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    prometheus.BuildFQName(_namespace, _subsystem, "taskrun_duration_seconds"),
		Help:    "Taskrun duration info",
		Buckets: buckets,
	}, []string{_environment, _template, _pipeline, _result, _task})

	_stepHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    prometheus.BuildFQName(_namespace, _subsystem, "step_duration_seconds"),
		Help:    "Step duration info",
		Buckets: buckets,
	}, []string{_environment, _template, _pipeline, _result, _task, _step})
}

func Observe(results *PipelineResults, data *global.HorizonMetaData) {
	if results == nil {
		return
	}
	prMetadata := results.Metadata
	prResult := results.PrResult
	trResults, stepResults := results.TrResults, results.StepResults

	_prHistogram.With(prometheus.Labels{
		_environment: data.Environment,
		_template:    data.Template,
		_pipeline:    prMetadata.Pipeline,
		_result:      prResult.Result,
	}).Observe(prResult.DurationSeconds)

	for _, trResult := range trResults {
		_trHistogram.With(prometheus.Labels{
			_environment: data.Environment,
			_template:    data.Template,
			_pipeline:    prMetadata.Pipeline,
			_task:        trResult.Task,
			_result:      trResult.Result,
		}).Observe(trResult.DurationSeconds)
	}

	for _, stepResult := range stepResults {
		_stepHistogram.With(prometheus.Labels{
			_environment: data.Environment,
			_template:    data.Template,
			_step:        stepResult.Step,
			_pipeline:    prMetadata.Pipeline,
			_task:        stepResult.Task,
			_result:      stepResult.Result,
		}).Observe(stepResult.DurationSeconds)
	}
}
