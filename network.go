// Copyright 2026 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package linkcable

import (
	"fmt"
	"sync"

	"github.com/moby/moby/v2/daemon/libnetwork/datastore"
)

// NetworksByID maps Docker network IDs to our network configuration and
// endpoint state information.
type NetworksByID map[string]*Network

// Network represents a (custom) Docker network managed by our driver: it stores
// both the persistent configuration data as well as purely in-memory state
// information.
type Network struct {
	cfg NetworkConfiguration

	m         sync.RWMutex
	endpoints []Endpoint
}

var _ Storable = (*Network)(nil)

// Persist writes the configuration of the passed network into the data store.
func (n *Network) Persist(store *datastore.Store) error {
	n.m.RLock()
	defer n.m.RUnlock()
	if err := store.PutObjectAtomic(&n.cfg); err != nil {
		return fmt.Errorf("cannot update network (ID=%q) configuration in data store, reason: %w",
			abbreviate(n.cfg.ID), err)
	}
	return nil
}

// Cease wipes the configuration of the passed network from the data store.
func (n *Network) Cease(store *datastore.Store) error {
	n.m.RLock()
	defer n.m.RUnlock()
	for attempt := 0; attempt < 10; attempt++ {
		switch err := store.DeleteObjectAtomic(&n.cfg); err {
		case nil:
			return nil
		case datastore.ErrKeyModified:
			var updatedNetwork NetworkConfiguration
			n.cfg.CopyTo(&updatedNetwork)
			if err := store.GetObject(&updatedNetwork); err != nil {
				return fmt.Errorf("cannot refresh network (ID=%q) configuration in order to delete; reason: %w",
					err)
			}
			updatedNetwork.CopyTo(&n.cfg)
			continue
		default:
			return fmt.Errorf("cannot remove network (ID=%q) configuration from data store, reason: %w",
				abbreviate(n.cfg.ID), err)
		}
	}
	return fmt.Errorf("too many failed attempts to remove network (ID=%q) configuration from data store",
		abbreviate(n.cfg.ID))
}
