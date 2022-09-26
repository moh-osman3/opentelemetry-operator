package prehooktargetfilter

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

type AllocatorPrehook interface {
	SetTargets(targets map[string]*allocation.TargetItem)
	TargetItems() map[string]*allocation.TargetItem
}

type AllocatorPrehookProvider func(log logr.Logger, allocator allocation.Allocator) AllocatorPrehook

var (
	registry = map[string]AllocatorPrehookProvider{}
)

func New(name string, log logr.Logger, allocator allocation.Allocator) (AllocatorPrehook, error) {
	if p, ok := registry[name]; ok {
		return p(log, allocator), nil
	}
	return nil, errors.New(fmt.Sprintf("unregistered filtering strategy: %s", name))
}

func Register(name string, provider AllocatorPrehookProvider) error {
	if _, ok := registry[name]; ok {
		return errors.New("already registered")
	}
	registry[name] = provider
	return nil
}

func init() {
	err := Register(noOpTargetFilterName, NewNoOpTargetFilter)
	if err != nil {
		panic(err)
	}
	err = Register(relabelConfigTargetFilterName, NewRelabelConfigTargetFilter)
	if err != nil {
		panic(err)
	}
}
