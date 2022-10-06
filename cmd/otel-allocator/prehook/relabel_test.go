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
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/assert"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
)

var (
	logger     = logf.Log.WithName("unit-tests")
	numTargets = 100

	RelabelConfigs = []RelabelConfigObj{
		{
			cfg: relabel.Config{
				Action:      "replace",
				Separator:   ";",
				Regex:       relabel.MustNewRegexp("(.*)"),
				Replacement: "$1",
			},
			isDrop: false,
		},
		{
			cfg: relabel.Config{
				SourceLabels: model.LabelNames{"i"},
				Regex:        relabel.MustNewRegexp("(.*)"),
				Action:       "keep",
			},
			isDrop: false,
		},
		{
			cfg: relabel.Config{
				SourceLabels: model.LabelNames{"i"},
				Regex:        relabel.MustNewRegexp("bad.*match"),
				Action:       "drop",
			},
			isDrop: false,
		},
		{
			// Note that usually a label not being present, with action=keep, would
			// mean the target is dropped. However it is hard to tell what labels will
			// never exist and which ones will exist after relabelling, so these targets
			// are kept in this step (will later be dealt with in Prometheus' relabel_config step)
			cfg: relabel.Config{
				SourceLabels: model.LabelNames{"label_not_present"},
				Regex:        relabel.MustNewRegexp("(.*)"),
				Action:       "keep",
			},
			isDrop: false,
		},
		{
			cfg: relabel.Config{
				SourceLabels: model.LabelNames{"i"},
				Regex:        relabel.MustNewRegexp("(.*)"),
				Action:       "drop",
			},
			isDrop: true,
		},
		{
			cfg: relabel.Config{
				SourceLabels: model.LabelNames{"collector"},
				Regex:        relabel.MustNewRegexp("(collector.*)"),
				Action:       "drop",
			},
			isDrop: true,
		},
		{
			cfg: relabel.Config{
				SourceLabels: model.LabelNames{"i"},
				Regex:        relabel.MustNewRegexp("bad.*match"),
				Action:       "keep",
			},
			isDrop: true,
		},
		{
			cfg: relabel.Config{
				SourceLabels: model.LabelNames{"collector"},
				Regex:        relabel.MustNewRegexp("collectors-n"),
				Action:       "keep",
			},
			isDrop: true,
		},
	}

	DefaultDropRelabelConfig = relabel.Config{
		SourceLabels: model.LabelNames{"i"},
		Regex:        relabel.MustNewRegexp("(.*)"),
		Action:       "drop",
	}
)

type RelabelConfigObj struct {
	cfg    relabel.Config
	isDrop bool
}

func colIndex(index, numCols int) int {
	if numCols == 0 {
		return -1
	}
	return index % numCols
}

func makeNNewTargets(n int, numCollectors int, startingIndex int) (map[string]*allocation.TargetItem, int) {
	toReturn := map[string]*allocation.TargetItem{}
	numItemsRemaining := n
	for i := startingIndex; i < n+startingIndex; i++ {
		collector := fmt.Sprintf("collector-%d", colIndex(i, numCollectors))
		label := model.LabelSet{
			"collector": model.LabelValue(collector),
			"i":         model.LabelValue(strconv.Itoa(i)),
			"total":     model.LabelValue(strconv.Itoa(n + startingIndex)),
		}
		newTarget := allocation.NewTargetItem(fmt.Sprintf("test-job-%d", i), "test-url", label, collector)
		// add a single replace, drop, or keep action as relabel_config for targets
		// index := rand.Intn(len(RelabelConfigs))
		var index int
		ind, _ := rand.Int(rand.Reader, big.NewInt(int64(len(RelabelConfigs))))

		index = int(ind.Int64())

		newTarget.RelabelConfigs = []*relabel.Config{
			&RelabelConfigs[index].cfg,
		}

		if RelabelConfigs[index].isDrop {
			numItemsRemaining--
		}
		toReturn[newTarget.Hash()] = newTarget
	}
	return toReturn, numItemsRemaining
}

func TestSetTargets(t *testing.T) {
	allocator := mockAllocator{targetItems: make(map[string]*allocation.TargetItem)}
	allocatorPrehook, err := New("relabel-config", logger, allocator)
	assert.Nil(t, err)

	targets, numRemaining := makeNNewTargets(numTargets, 3, 0)
	allocatorPrehook.SetTargets(targets)
	remainingTargetItems := allocatorPrehook.TargetItems()
	assert.Len(t, remainingTargetItems, numRemaining)
}
