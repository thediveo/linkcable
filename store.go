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
	"log/slog"

	"github.com/moby/moby/v2/daemon/libnetwork/datastore"
	"github.com/thediveo/nonstd/xslog"
)

const (
	KVPrefix         = "linkcable"
	KVNetworkPrefix  = KVPrefix + "network"
	KVEndpointPrefix = KVPrefix + "endpoint"
)

// storeUpdate writes the passed network configuration into the persistence data
// store.
func (d *Driver) storeUpdate(n *Network) error {
	n.m.RLock()
	defer n.m.RUnlock()
	if err := d.store.PutObjectAtomic(&n.cfg); err != nil {
		return fmt.Errorf("cannot update network configuration (ID=%q) in data store, reason: %w",
			n.cfg.ID, err)
	}
	return nil
}

func (d *Driver) storeDelete(n *Network) error {
	if d.store == nil {
		return nil
	}
	return nil // FIXME:
}

// restoreNetworks restores the in-memory plugin state from the persistent
// network configuration in the data store.
func (d *Driver) restoreNetworks(tabluarasa bool) error {
	networks, err := d.store.List(&NetworkConfiguration{})
	if err != nil {
		if err != datastore.ErrKeyNotFound {
			return fmt.Errorf("cannot retrieve network configuration from data store; reason: %w", err)
		}
		slog.Debug("no network configuration found")
		return nil
	}
	for _, netw := range networks {
		network := &Network{}
		netw.CopyTo(&network.cfg)
		if tabluarasa {
			slog.Warn("forgetting network configuration",
				slog.String("id", network.cfg.ID))
			if err := d.storeDelete(network); err != nil {
				slog.Error("cannot delete network configuration from data store",
					slog.String("id", network.cfg.ID),
					xslog.Error(err))
			}
			continue
		}
		slog.Debug("restoring network configuration",
			slog.String("id", network.cfg.ID))
		// TODO: restore
	}
	return nil
}
