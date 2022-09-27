package prehooktargetfilter

import (
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

func NewRelabelConfigTargetFilter(log logr.Logger, allocator allocation.Allocator) AllocatorPrehook {
	return &RelabelConfigTargetFilter{
		log:        log,
		allocator:  allocator,
	}
}

func (tf *RelabelConfigTargetFilter) SetTargets(targets map[string]*allocation.TargetItem) {
	filteredTargets := make(map[string]*allocation.TargetItem)
	for jobName, tItem := range targets {
		keepTarget := true
		for _, cfg := range tItem.RelabelConfigs {
			if IsDropTarget(tItem.Label, cfg) {
				keepTarget = false
				break
			}
		}

		if keepTarget {
			filteredTargets[jobName] = tItem
		}
	}

	tf.allocator.SetTargets(filteredTargets)
	tf.targetItems = filteredTargets
	return
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
// be dropped or not - function should be called for each item in the relabel_config
// TODO: add more actions?
func IsDropTarget(lset model.LabelSet, cfg *relabel.Config) bool {
	values := make([]string, 0, len(cfg.SourceLabels))
	for _, ln := range cfg.SourceLabels {
		if val, ok := lset[ln]; ok {
			values = append(values, string(val))
		}
	}
	val := strings.Join(values, cfg.Separator)

	switch cfg.Action {
	case "drop":
		if cfg.Regex.MatchString(val) {
			return true
		}
	case "keep":
		if !cfg.Regex.MatchString(val) {
			return true
		}
	default:
		return false
	}

	return false
}