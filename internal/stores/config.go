package stores

import (
	"sync"

	addonsv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
)

type ConfigStore struct {
	sync.RWMutex
	config addonsv1alpha1.ArgoAddonSpec
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		config: addonsv1alpha1.ArgoAddonSpec{},
	}
}

func (s *ConfigStore) Get() addonsv1alpha1.ArgoAddonSpec {
	s.RLock()
	defer s.RUnlock()
	return s.config
}

func (s *ConfigStore) Update(config addonsv1alpha1.ArgoAddonSpec) {
	s.Lock()
	defer s.Unlock()
	s.config = config
}
