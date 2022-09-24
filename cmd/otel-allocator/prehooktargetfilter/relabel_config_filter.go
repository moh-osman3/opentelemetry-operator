package prehooktargetfilter

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
)

type RelabelConfigTargetFilter struct {
	log         logr.Logger       
	targetItems map[string]*allocation.TargetItem 
	allocator   *allocation.Allocator
}

func NewRelabelConfigTargetFilter(log logr.Logger, allocator *allocation.Allocator) AllocatorPrehook {
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
			if !Relabel(tItem.Label, cfg) {
				keepTarget = false
				break
			}
		}

		if keepTarget {
			filteredTargets[jobName] = tItem
		}
	}

	(*tf.allocator).SetTargets(filteredTargets)
	return
}

// Goal of this function is to determine whether a given target should
// be dropped or not - function should be called for each item in the relabel_config
// TODO: add more actions?
func Relabel(lset model.LabelSet, cfg *relabel.Config) bool {
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
			return false
		}
	default:
		return true
	}

	return true
}