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
	"encoding/json"
	"log/slog"

	"github.com/docker/go-plugins-helpers/network"
	"github.com/moby/moby/v2/daemon/libnetwork/datastore"
	"github.com/thediveo/nonstd/xslog"
)

// NetworkConfiguration contains the configuration information of a custom
// Docker network that needs to be persisted in a KV data store.
type NetworkConfiguration struct {
	ID      string // ID assigned by Docker/Moby to this network.
	Options map[string]any
	IPAMsv4 []*network.IPAMData
	IPAMsv6 []*network.IPAMData

	dbExists bool
	dbIndex  uint64
}

var _ datastore.KVObject = (*NetworkConfiguration)(nil)

// Key to be used in the KV store.
func (n *NetworkConfiguration) Key() []string {
	return []string{KVNetworkPrefix, n.ID}
}

// KeyPrefix is the immediate parent key for use in tree walks.
func (n *NetworkConfiguration) KeyPrefix() []string {
	return []string{KVNetworkPrefix}
}

// Value marshals the persistent network configuration as JSON into a byte slice
// to be stored in the KV store. In case of a marshalling error, it returns a
// nil slice.
func (n *NetworkConfiguration) Value() []byte {
	b, err := json.Marshal(n)
	if err != nil {
		slog.Error("cannot marshal network configuration data",
			slog.String("network", n.ID),
			xslog.Error(err))
		return nil
	}
	return b
}

// SetValue is used by the KV store to set the object's value when loaded from
// the data store.
func (n *NetworkConfiguration) SetValue(v []byte) error {
	return json.Unmarshal(v, n)
}

// Index returns the latest DB index as seen by this configuration object.
func (n *NetworkConfiguration) Index() uint64 { return n.dbIndex }

// SetIndex allows the data store to set the latest DB Index for this
// configuration object.
func (n *NetworkConfiguration) SetIndex(i uint64) {
	n.dbExists = true
	n.dbIndex = i
}

// Exists returns true if this configuration object exists in the data store;
// otherwise false if it hasn't been stored yet. When
// [NetworkConfiguration.SetIndex] is called, the object has been stored.
func (n *NetworkConfiguration) Exists() bool { return n.dbExists }

// Skip returning true, allows a KV object avoiding getting persisted in the
// data store.
func (n *NetworkConfiguration) Skip() bool { return false }

// New returns a new zero-value configuration object. Please note that
// [KVObject's New comment] could be read as behaving somewhat clone-ish, but
// libnetwork's internal drivers, such as macvlan, show that this really is
// new-ish returning a new zero value configuration object.
//
// [KVObject's New comment]: https://pkg.go.dev/github.com/moby/moby/v2/daemon/libnetwork/datastore#KVObject
func (n *NetworkConfiguration) New() datastore.KVObject {
	return &NetworkConfiguration{}
}

// CopyTo deep copies the configuration of this object into the passed
// destination object.
func (n *NetworkConfiguration) CopyTo(other datastore.KVObject) error {
	*(other.(*NetworkConfiguration)) = *n // shallow copy
	return nil
}
