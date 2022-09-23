package targetfilter

import (
//	"errors"
//	"fmt"
//	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
)

type TargetFilter struct {
	log        logr.Logger       
	close      chan struct{}
	sig        chan int
	targets    map[string]*allocation.TargetItem 
}

func NewTargetFilter(log logr.Logger) *TargetFilter {
	return &TargetFilter{
		log:        log,
		close:      make(chan struct{}),
		sig:        make(chan int),
	}
}


func (tf *TargetFilter) Watch(fn func(targets map[string]*allocation.TargetItem)) {
	log := tf.log.WithValues("component", "opentelemetry-targetallocator")
	go func() {
		for {
			select {
			case <- tf.close:
				log.Info("Service Discovery watch event stopped: discovery manager closed")
				return
			case <- tf.sig:
				if tf.targets != nil {
					fn(tf.targets)
				}
			}
		}
	}()
}

func (tf *TargetFilter) Close() {
	close(tf.close)
}

func (tf *TargetFilter) RelabelConfigFilterTargets(targets map[string]*allocation.TargetItem) {
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

	tf.sig <- 1
	tf.targets = filteredTargets
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
