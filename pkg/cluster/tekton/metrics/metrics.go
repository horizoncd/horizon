package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	_prHistogram   *prometheus.HistogramVec
	_trHistogram   *prometheus.HistogramVec
	_stepHistogram *prometheus.HistogramVec
)

const (
	_application = "application"
	_cluster     = "cluster"
	_environment = "environment"
	_name        = "name"
	_pipeline    = "pipeline"
	_task        = "task"
	_result      = "result"
)

func init() {
	_prHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gitops_pipelinerun_duration_seconds",
		Help:    "PipelineRun duration info",
		Buckets: append([]float64{0}, prometheus.ExponentialBuckets(1, 2, 12)...),
	}, []string{_application, _cluster, _environment, _pipeline, _result})

	_trHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gitops_taskrun_duration_seconds",
		Help:    "Taskrun duration info",
		Buckets: append([]float64{0}, prometheus.ExponentialBuckets(1, 2, 12)...),
	}, []string{_application, _cluster, _environment, _pipeline, _task, _result})

	_stepHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gitops_step_duration_seconds",
		Help:    "Step duration info",
		Buckets: append([]float64{0}, prometheus.ExponentialBuckets(1, 2, 12)...),
	}, []string{_application, _cluster, _environment,
		_name, _pipeline, _task, _result})
}

func Observe(wpr *WrappedPipelineRun) {
	if wpr == nil || wpr.PipelineRun == nil {
		return
	}
	prMetadata := wpr.ResolveMetadata()
	prBusinessData := wpr.ResolveBusinessData()
	prResult := wpr.ResolvePrResult()
	trResults, stepResults := wpr.ResolveTrAndStepResults()

	_prHistogram.With(prometheus.Labels{
		_application: prBusinessData.Application,
		_cluster:     prBusinessData.Cluster,
		_environment: prBusinessData.Environment,
		_pipeline:    prMetadata.Pipeline,
		_result:      prResult.Result.String(),
	}).Observe(prResult.DurationSeconds)

	for _, trResult := range trResults {
		_trHistogram.With(prometheus.Labels{
			_application: prBusinessData.Application,
			_cluster:     prBusinessData.Cluster,
			_environment: prBusinessData.Environment,
			_pipeline:    prMetadata.Pipeline,
			_task:        trResult.Task,
			_result:      trResult.Result.String(),
		}).Observe(trResult.DurationSeconds)
	}

	for _, stepResult := range stepResults {
		_stepHistogram.With(prometheus.Labels{
			_application: prBusinessData.Application,
			_cluster:     prBusinessData.Cluster,
			_environment: prBusinessData.Environment,
			_name:        stepResult.Name,
			_pipeline:    prMetadata.Pipeline,
			_task:        stepResult.Task,
			_result:      stepResult.Result.String(),
		}).Observe(stepResult.DurationSeconds)
	}
}
