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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
)

var _ allocation.Allocator = &mockAllocator{}

type mockAllocator struct {
	targetItems map[string]*allocation.TargetItem
}

func (allocator mockAllocator) SetTargets(targets map[string]*allocation.TargetItem) {
	for k, v := range targets {
		allocator.targetItems[k] = v
	}
}

func (allocator mockAllocator) TargetItems() map[string]*allocation.TargetItem {
	return allocator.targetItems
}

func (allocator mockAllocator) SetCollectors(collectors map[string]*allocation.Collector) {
}
func (allocator mockAllocator) Collectors() map[string]*allocation.Collector {
	return nil
}

func TestNoOpSetTargets(t *testing.T) {
	allocator := mockAllocator{targetItems: make(map[string]*allocation.TargetItem)}

	allocatorPrehook, err := New("no-op", logger, allocator)
	assert.Nil(t, err)

	targets, _, _ := makeNNewTargets(numTargets, 3, 0)
	allocatorPrehook.SetTargets(targets)
	remainingTargetItems := allocatorPrehook.TargetItems()
	assert.Len(t, remainingTargetItems, numTargets)
	assert.Equal(t, remainingTargetItems, targets)
}
