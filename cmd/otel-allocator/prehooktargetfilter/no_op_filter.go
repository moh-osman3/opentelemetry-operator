package prehooktargetfilter

import (
	"github.com/go-logr/logr"
	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
)

const relabelConfigTargetFilter = "relabel_config_filter"

type NoOpTargetFilter struct {
	log         logr.Logger       
	allocator   *allocation.Allocator
}

func NewNoOpTargetFilter(log logr.Logger, allocator *allocation.Allocator) AllocatorPrehook {
	return &NoOpTargetFilter{
		log:        log,
		allocator:  allocator,
	}
}

func (tf *NoOpTargetFilter) SetTargets(targets map[string]*allocation.TargetItem) {
	(*tf.allocator).SetTargets(targets)
	return
}