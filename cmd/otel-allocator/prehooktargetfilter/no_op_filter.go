package prehooktargetfilter

import (
	"sync"

	"github.com/go-logr/logr"
	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
)

const relabelConfigTargetFilter = "relabel_config_filter"

type NoOpTargetFilter struct {
	m           sync.RWMutex
	log         logr.Logger       
	targetItems map[string]*allocation.TargetItem
	allocator   allocation.Allocator
}

func NewNoOpTargetFilter(log logr.Logger, allocator allocation.Allocator) AllocatorPrehook {
	return &NoOpTargetFilter{
		log:        log,
		allocator:  allocator,
	}
}

func (tf *NoOpTargetFilter) SetTargets(targets map[string]*allocation.TargetItem) {
	tf.allocator.SetTargets(targets)
	return
}

// TargetItems returns a shallow copy of the targetItems map.
func (c *NoOpTargetFilter) TargetItems() map[string]*allocation.TargetItem {
	c.m.RLock()
	defer c.m.RUnlock()
	targetItemsCopy := make(map[string]*allocation.TargetItem)
	for k, v := range c.targetItems {
		targetItemsCopy[k] = v
	}
	return targetItemsCopy
}