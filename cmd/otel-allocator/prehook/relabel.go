// Copyright The OpenTelemetry Authors
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

package prehook

import (
	"github.com/go-logr/logr"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/prehook"
)

type RelabelConfigTargetFilter struct {
	log        logr.Logger
	filterFunc FilterFunc
}

func NewRelabelConfigTargetFilter(log logr.Logger) Hook {
	return &RelabelConfigTargetFilter{
		log:             log,
	}
}

// helper function converts from model.LabelSet to []labels.Label
func ConvertLabelToPromLabelSet(lbls model.LabelSet) []labels.Label {
	newLabels := make([]labels.Label, len(lbls))
	index := 0
	for k,v := range lbls {
		newLabels[index].Name = string(k)
		newLabels[index].Value = string(v)
		index++
	}
	return newLabels
}

func (tf *RelabelConfigTargetFilter) Apply(targets map[string]*allocation.TargetItem) map[string]*allocation.TargetItem {
	numTargets := len(targets)
	numRemainingTargets := numTargets
	for jobName, tItem := range targets {
		keepTarget := true
		lset := ConvertLabelToPromLabelSet(tItem.Label)
		for _, cfg := range tItem.RelabelConfigs {
			if new_lset := relabel.Process(lset, cfg); new_lset == nil {
				keepTarget = false
				break // inner loop
			} else {
				lset = new_lset
			}
		}

		if !keepTarget {
			delete(targets, jobName)
			numRemainingTargets--
		}
	}

	// tf.allocator.SetTargets(targets)
	tf.log.V(2).Info("Filtering complete", "seen", numTargets, "kept", numRemainingTargets)
}

func init() {
	err := Register(relabelConfigTargetFilterName, NewRelabelConfigTargetFilter)
	if err != nil {
		panic(err)
	}
}