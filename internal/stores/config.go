// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"sync"

	addonsv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
)

type ConfigStore struct {
	sync.RWMutex
	config *addonsv1alpha1.ArgoAddonSpec
	notify chan struct{}
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		config: &addonsv1alpha1.ArgoAddonSpec{},
		notify: make(chan struct{}, 1),
	}
}

func (s *ConfigStore) Get() *addonsv1alpha1.ArgoAddonSpec {
	s.RLock()
	defer s.RUnlock()

	return s.config
}

func (s *ConfigStore) Update(config *addonsv1alpha1.ArgoAddonSpec) {
	s.Lock()
	defer s.Unlock()
	s.config = config

	s.notify <- struct{}{}
}

func (s *ConfigStore) NotifyChannel() <-chan struct{} {
	return s.notify
}
