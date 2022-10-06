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
	"sync"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
)

type RelabelConfigTargetFilter struct {
	m           sync.RWMutex
	log         logr.Logger
	targetItems map[string]*allocation.TargetItem
	allocator   allocation.Allocator
}

func NewRelabelConfigTargetFilter(log logr.Logger, allocator allocation.Allocator) Hook {
	return &RelabelConfigTargetFilter{
		log:         log,
		allocator:   allocator,
		targetItems: make(map[string]*allocation.TargetItem),
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

func (tf *RelabelConfigTargetFilter) SetTargets(targets map[string]*allocation.TargetItem) {
	numRemainingTargets := 0
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

		if keepTarget {
			tf.targetItems[jobName] = tItem
			numRemainingTargets++
		}
	}

	tf.allocator.SetTargets(tf.targetItems)
	tf.log.V(2).Info("Filtering complete", "seen", len(targets), "kept", numRemainingTargets)
}

// TargetItems returns a shallow copy of the targetItems map.
func (tf *RelabelConfigTargetFilter) TargetItems() map[string]*allocation.TargetItem {
	return tf.allocator.TargetItems()
}

func init() {
	err := Register(relabelConfigTargetFilterName, NewRelabelConfigTargetFilter)
	if err != nil {
		panic(err)
	}
}