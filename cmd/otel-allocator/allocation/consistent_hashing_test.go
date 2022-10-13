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

package allocation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/prometheus/prometheus/config"

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/prehook"
	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/targetscommon"
)

var _ prehook.Hook = &mockHook{}

type mockHook struct {
}

func (allocator mockHook) Apply(targets map[string]*targetscommon.TargetItem) map[string]*targetscommon.TargetItem {
	return targets
}

func (allocator mockHook) SetConfig(map[string]*config.Config) {
}

func TestCanSetSingleTarget(t *testing.T) {
	cols := makeNCollectors(3, 0)
	hook := mockHook{}
	c := newConsistentHashingAllocator(logger, hook)
	c.SetCollectors(cols)
	c.SetTargets(makeNNewTargets(1, 3, 0))
	actualTargetItems := c.TargetItems()
	assert.Len(t, actualTargetItems, 1)
	for _, item := range actualTargetItems {
		assert.Equal(t, "collector-2", item.CollectorName)
	}
}

func TestRelativelyEvenDistribution(t *testing.T) {
	numCols := 15
	numItems := 10000
	cols := makeNCollectors(numCols, 0)
	var expectedPerCollector = float64(numItems / numCols)
	expectedDelta := (expectedPerCollector * 1.5) - expectedPerCollector
	hook := mockHook{}
	c := newConsistentHashingAllocator(logger, hook)
	c.SetCollectors(cols)
	c.SetTargets(makeNNewTargets(numItems, 0, 0))
	actualTargetItems := c.TargetItems()
	assert.Len(t, actualTargetItems, numItems)
	actualCollectors := c.Collectors()
	assert.Len(t, actualCollectors, numCols)
	for _, col := range actualCollectors {
		t.Logf("col: %s \ttargets: %d", col.Name, col.NumTargets)
		assert.InDelta(t, col.NumTargets, expectedPerCollector, expectedDelta)
	}
}

func TestFullReallocation(t *testing.T) {
	cols := makeNCollectors(10, 0)
	hook := mockHook{}
	c := newConsistentHashingAllocator(logger, hook)
	c.SetCollectors(cols)
	c.SetTargets(makeNNewTargets(10000, 10, 0))
	actualTargetItems := c.TargetItems()
	assert.Len(t, actualTargetItems, 10000)
	actualCollectors := c.Collectors()
	assert.Len(t, actualCollectors, 10)
	newCols := makeNCollectors(10, 10)
	c.SetCollectors(newCols)
	updatedTargetItems := c.TargetItems()
	assert.Len(t, updatedTargetItems, 10000)
	updatedCollectors := c.Collectors()
	assert.Len(t, updatedCollectors, 10)
	for _, item := range updatedTargetItems {
		_, ok := updatedCollectors[item.CollectorName]
		assert.True(t, ok, "Some items weren't reallocated correctly")
	}
}

func TestNumRemapped(t *testing.T) {
	numItems := 10_000
	numInitialCols := 15
	numFinalCols := 16
	expectedDelta := float64((numFinalCols - numInitialCols) * (numItems / numFinalCols))
	cols := makeNCollectors(numInitialCols, 0)
	hook := mockHook{}
	c := newConsistentHashingAllocator(logger, hook)
	c.SetCollectors(cols)
	c.SetTargets(makeNNewTargets(numItems, numInitialCols, 0))
	actualTargetItems := c.TargetItems()
	assert.Len(t, actualTargetItems, numItems)
	actualCollectors := c.Collectors()
	assert.Len(t, actualCollectors, numInitialCols)
	newCols := makeNCollectors(numFinalCols, 0)
	c.SetCollectors(newCols)
	updatedTargetItems := c.TargetItems()
	assert.Len(t, updatedTargetItems, numItems)
	updatedCollectors := c.Collectors()
	assert.Len(t, updatedCollectors, numFinalCols)
	countRemapped := 0
	countNotRemapped := 0
	for _, item := range updatedTargetItems {
		previousItem, ok := actualTargetItems[item.Hash()]
		assert.True(t, ok)
		if previousItem.CollectorName != item.CollectorName {
			countRemapped++
		} else {
			countNotRemapped++
		}
	}
	assert.InDelta(t, numItems/numFinalCols, countRemapped, expectedDelta)
}
