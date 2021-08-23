// pmm-agent
// Copyright 2019 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package lock implements simple locking service that allows to actions and jobs lock required entities.
package lock

import (
	"sync"
)

// Entity represents something that should be used exclusively by job/action/process.
// For example, it can be some directory or tool.
type Entity string

const (
	PBM = "pbm"
)

// Service is locking service. It allows acquiring of exclusive locks for entities.
type Service struct {
	m        sync.Mutex
	entities map[Entity]struct{}
}

// New creates new locks service.
func New() *Service {
	return &Service{
		entities: make(map[Entity]struct{}),
	}
}

// TryAcquire checks that passed entities can be locked. If all passed entities available for locking
// returns ture, otherwise returns false. This method locks all or nothing.
func (r *Service) TryAcquire(entities ...Entity) bool {
	if len(entities) == 0 {
		return true
	}

	r.m.Lock()
	defer r.m.Unlock()

	// check that there is no active locks for passed entities
	for _, entity := range entities {
		if _, ok := r.entities[entity]; ok {
			return false
		}
	}

	for _, entity := range entities {
		r.entities[entity] = struct{}{}
	}
	return true
}

// Release releases locks for passed entities.
func (r *Service) Release(entities ...Entity) {
	if len(entities) == 0 {
		return
	}

	r.m.Lock()
	defer r.m.Unlock()

	for _, entity := range entities {
		delete(r.entities, entity)
	}
}
