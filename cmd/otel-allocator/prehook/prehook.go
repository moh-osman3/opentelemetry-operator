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

	"github.com/open-telemetry/opentelemetry-operator/cmd/otel-allocator/allocation"
)

const (
	noOpTargetFilterName          = "no-op"
	relabelConfigTargetFilterName = "relabel-config"
)

type Hook interface {
	SetTargets(targets map[string]*allocation.TargetItem)
	TargetItems() map[string]*allocation.TargetItem
}

type Provider func(log logr.Logger, allocator allocation.Allocator) Hook

var (
	registry = map[string]Provider{}
)

func New(name string, log logr.Logger, allocator allocation.Allocator) (Hook, error) {
	if p, ok := registry[name]; ok {
		return p(log.WithName("Prehook").WithName(name), allocator), nil
	}
	return nil, fmt.Errorf("unregistered filtering strategy: %s", name)
}

func Register(name string, provider Provider) error {
	if _, ok := registry[name]; ok {
		return errors.New("already registered")
	}
	registry[name] = provider
	return nil
}