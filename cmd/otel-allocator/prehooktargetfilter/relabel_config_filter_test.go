package prehooktargetfilter

import (
	"testing"
	"fmt"
	"strconv"

	"github.com/stretchr/testify/assert"
	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	logger = logf.Log.WithName("unit-tests")

	DefaultRelabelConfig = relabel.Config{
		Action:      "replace",
		Separator:   ";",
		Regex:       relabel.MustNewRegexp("(.*)"),
		Replacement: "$1",
	}

	DefaultDropRelabelConfig = relabel.Config{
		SourceLabels: model.LabelNames{"a"},
		Regex:        relabel.MustNewRegexp("(.*)"),
		Action:       "drop",	
	}

	DefaultKeepRelabelConfig = relabel.Config{
			SourceLabels: model.LabelNames{"a"},
			Regex:        relabel.MustNewRegexp("bad.*match"),
			Action:       "keep",
	}
)

func colIndex(index, numCols int) int {
	if numCols == 0 {
		return -1
	}
	return index % numCols
}

func makeNNewTargets(n int, numCollectors int, startingIndex int) map[string]*allocation.TargetItem {
	toReturn := map[string]*allocation.TargetItem{}
	for i := startingIndex; i < n+startingIndex; i++ {
		collector := fmt.Sprintf("collector-%d", colIndex(i, numCollectors))
		label := model.LabelSet{
			"collector": model.LabelValue(collector),
			"i":         model.LabelValue(strconv.Itoa(i)),
			"total":     model.LabelValue(strconv.Itoa(n + startingIndex)),
		}
		newTarget := allocation.NewTargetItem(fmt.Sprintf("test-job-%d", i), "test-url", label, collector)
		// add a single replace or drop action as relabel_config for targets
		if i % 2 == 0 {
			newTarget.RelabelConfigs = []*relabel.Config{
				&DefaultRelabelConfig,
			}
		} else {
			newTarget.RelabelConfigs = []*relabel.Config{
				&DefaultKeepRelabelConfig,
			}
		}
		toReturn[newTarget.Hash()] = newTarget
	}
	return toReturn
}

func TestIsDroppedTarget(t *testing.T) {
	tests := []struct {
		input   model.LabelSet 
		relabel *relabel.Config
		output  bool
	}{
		{
			input: model.LabelSet{
				"a": "foo",
				"b": "bar",
			},
			relabel: &DefaultDropRelabelConfig,
			output: true,
		},
	}
	for _, test := range tests {
		res := IsDropTarget(test.input, test.relabel)
		assert.Equal(t, test.output, res)
	}
}

func TestSetTargets(t *testing.T) {
	allocator, err := allocation.New("least-weighted", logger)
	assert.Nil(t, err)

	allocatorPrehook, err := New("relabel-config", logger, allocator)
	assert.Nil(t, err)

	targets := makeNNewTargets(100, 3, 0)
	allocatorPrehook.SetTargets(targets)
	remainingTargetItems := allocatorPrehook.TargetItems()
	assert.Len(t, remainingTargetItems, 50)
}