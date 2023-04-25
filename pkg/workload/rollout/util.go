// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rollout

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/fnv"

	rolloutsv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func computeStepHash(rollout *rolloutsv1alpha1.Rollout) string {
	if rollout.Spec.Strategy.BlueGreen != nil || rollout.Spec.Strategy.Canary == nil {
		return ""
	}
	rolloutStepHasher := fnv.New32a()
	stepsBytes, err := json.Marshal(rollout.Spec.Strategy.Canary.Steps)
	if err != nil {
		panic(err)
	}
	_, err = rolloutStepHasher.Write(stepsBytes)
	if err != nil {
		panic(err)
	}
	return rand.SafeEncodeString(fmt.Sprint(rolloutStepHasher.Sum32()))
}

type podHashFields struct {
	InitContainers    []corev1.Container
	Containers        []corev1.Container
	RestartPolicy     corev1.RestartPolicy
	Affinity          *corev1.Affinity
	Tolerations       []corev1.Toleration
	PriorityClassName string
}

func assignContainers(containers []corev1.Container) []corev1.Container {
	ctrs := make([]corev1.Container, 0, len(containers))
	for _, container := range containers {
		ctr := corev1.Container{
			Name:       container.Name,
			Image:      container.Image,
			Command:    container.Command,
			Args:       container.Args,
			WorkingDir: container.WorkingDir,
		}
		for _, env := range container.Env {
			ctr.Env = append(ctr.Env, corev1.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			})
		}
		ctrs = append(ctrs, ctr)
	}
	return ctrs
}

func computePodSpecHash(spec corev1.PodSpec) string {
	fields := podHashFields{
		InitContainers: assignContainers(spec.InitContainers),
		Containers:     assignContainers(spec.Containers),
	}
	rolloutSpecHasher := fnv.New32a()
	_ = gob.NewEncoder(rolloutSpecHasher).Encode(fields)
	return rand.SafeEncodeString(fmt.Sprint(rolloutSpecHasher.Sum32()))
}
