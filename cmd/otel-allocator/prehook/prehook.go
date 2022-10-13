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
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/prometheus/prometheus/config"

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/targetscommon"
)

const (
	relabelConfigTargetFilterName = "relabel-config"
)

type Hook interface {
	Apply(map[string]*targetscommon.TargetItem) map[string]*targetscommon.TargetItem
	SetConfig(map[string]*config.Config)
}

type HookProvider func(log logr.Logger) Hook

var (
	registry = map[string]HookProvider{}
)

func New(name string, log logr.Logger) (Hook, error) {
	if p, ok := registry[name]; ok {
		return p(log.WithName("Prehook").WithName(name)), nil
	}
	return nil, fmt.Errorf("unregistered filtering strategy: %s", name)
}

func Register(name string, provider HookProvider) error {
	if _, ok := registry[name]; ok {
		return errors.New("already registered")
	}
	registry[name] = provider
	return nil
}