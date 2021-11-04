// fork from https://github.com/tektoncd/cli/blob/v0.13.1/pkg/log/pipeline_reader.go

package log

import (
	"fmt"

	"github.com/tektoncd/cli/pkg/pipeline"
	"github.com/tektoncd/cli/pkg/pipelinerun"
	trh "github.com/tektoncd/cli/pkg/taskrun"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reader) readPipelineLog() (<-chan Log, <-chan error, error) {
	pr, err := pipelinerun.GetV1beta1(r.clients, r.run, metav1.GetOptions{}, r.ns)
	if err != nil {
		return nil, nil, err
	}
	return r.readAvailablePipelineLogs(pr)
}

func (r *Reader) readAvailablePipelineLogs(pr *v1beta1.PipelineRun) (<-chan Log, <-chan error, error) {
	ordered, err := r.getOrderedTasks(pr)
	if err != nil {
		return nil, nil, err
	}

	taskRuns := trh.Filter(ordered, r.tasks)

	logC := make(chan Log)
	errC := make(chan error)

	go func() {
		defer close(logC)
		defer close(errC)

		// clone the object to keep task number and name separately
		c := r.clone()
		for i, tr := range taskRuns {
			c.setUpTask(i+1, tr)
			c.pipeLogs(logC, errC)
		}

		if !empty(pr.Status) && pr.Status.Conditions[0].Status == corev1.ConditionFalse {
			errC <- fmt.Errorf(pr.Status.Conditions[0].Message)
		}
	}()

	return logC, errC, nil
}

func (r *Reader) pipeLogs(logC chan<- Log, errC chan<- error) {
	tlogC, terrC, err := r.readTaskLog()
	if err != nil {
		errC <- err
		return
	}

	for tlogC != nil || terrC != nil {
		select {
		case l, ok := <-tlogC:
			if !ok {
				tlogC = nil
				continue
			}
			logC <- Log{Task: l.Task, Step: l.Step, Log: l.Log}

		case e, ok := <-terrC:
			if !ok {
				terrC = nil
				continue
			}
			errC <- fmt.Errorf("failed to get logs for task %s : %s", r.task, e)
		}
	}
}

func (r *Reader) setUpTask(taskNumber int, tr trh.Run) {
	r.setNumber(taskNumber)
	r.setRun(tr.Name)
	r.setTask(tr.Task)
}

// getOrderedTasks get Tasks in order from Spec.PipelineRef or Spec.PipelineSpec
// and return trh.Run after converted taskruns into trh.Run.
func (r *Reader) getOrderedTasks(pr *v1beta1.PipelineRun) ([]trh.Run, error) {
	var tasks []v1beta1.PipelineTask

	switch {
	case pr.Spec.PipelineRef != nil:
		pl, err := pipeline.GetV1beta1(r.clients, pr.Spec.PipelineRef.Name, metav1.GetOptions{}, r.ns)
		if err != nil {
			return nil, err
		}
		tasks = pl.Spec.Tasks
		tasks = append(tasks, pl.Spec.Finally...)
	case pr.Spec.PipelineSpec != nil:
		tasks = pr.Spec.PipelineSpec.Tasks
		tasks = append(tasks, pr.Spec.PipelineSpec.Finally...)
	default:
		return nil, fmt.Errorf("pipelinerun %s did not provide PipelineRef or PipelineSpec", pr.Name)
	}

	// Sort taskruns, to display the taskrun logs as per pipeline tasks order
	return trh.SortTasksBySpecOrder(tasks, pr.Status.TaskRuns), nil
}

func empty(status v1beta1.PipelineRunStatus) bool {
	if status.Conditions == nil {
		return true
	}
	return len(status.Conditions) == 0
}
