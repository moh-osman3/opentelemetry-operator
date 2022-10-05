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
	"fmt"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
)

type RelabelConfigTargetFilter struct {
	m           sync.RWMutex
	log         logr.Logger
	targetItems map[string]*allocation.TargetItem
	allocator   allocation.Allocator
}

func NewRelabelConfigTargetFilter(log logr.Logger, allocator allocation.Allocator) Prehook {
	return &RelabelConfigTargetFilter{
		log:         log,
		allocator:   allocator,
		targetItems: make(map[string]*allocation.TargetItem),
	}
}

func (tf *RelabelConfigTargetFilter) SetTargets(targets map[string]*allocation.TargetItem) {
	numRemainingTargets := 0
	for jobName, tItem := range targets {
		keepTarget := true
		for _, cfg := range tItem.RelabelConfigs {
			if tf.IsDropTarget(tItem.Label, cfg) {
				keepTarget = false
				break
			}
		}

		if keepTarget {
			tf.targetItems[jobName] = tItem
			numRemainingTargets++
		}
	}

	tf.allocator.SetTargets(tf.targetItems)
	tf.log.Info(fmt.Sprintf("Relabel filtering completed. Keeping %d target(s) out of %d total targets", numRemainingTargets, len(targets)))
}

// TargetItems returns a shallow copy of the targetItems map.
func (tf *RelabelConfigTargetFilter) TargetItems() map[string]*allocation.TargetItem {
	tf.m.RLock()
	defer tf.m.RUnlock()
	targetItemsCopy := make(map[string]*allocation.TargetItem)
	for k, v := range tf.targetItems {
		targetItemsCopy[k] = v
	}
	return targetItemsCopy
}

// Goal of this function is to determine whether a given target should
// be dropped or not - function should be called for each item in the relabel_config.
func (tf *RelabelConfigTargetFilter) IsDropTarget(lset model.LabelSet, cfg *relabel.Config) bool {
	values := make([]string, 0, len(cfg.SourceLabels))
	for _, ln := range cfg.SourceLabels {
		if val, ok := lset[ln]; ok {
			values = append(values, string(val))
		} else {
			tf.log.Info(fmt.Sprintf("label %v not found in lset, skipping..", ln))
			return false
		}
	}
	val := strings.Join(values, cfg.Separator)

	if cfg.Action == "drop" {
		if cfg.Regex.MatchString(val) {
			return true
		}
	} else if cfg.Action == "keep" {
		if !cfg.Regex.MatchString(val) {
			return true
		}
	}

	return false
}

func init() {
	err := Register(relabelConfigTargetFilterName, NewRelabelConfigTargetFilter)
	if err != nil {
		panic(err)
	}
}