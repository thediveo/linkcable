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
	"net"

	"github.com/moby/moby/v2/daemon/libnetwork/datastore"
	"github.com/thediveo/nonstd/xslog"
)

// EndpointConfiguration contains the configuration information of an CNM
// endpoint that connects a container to a network (or vice versa).
type EndpointConfiguration struct {
	ID        string
	NetworkID string
	Ifname    string
	MAC       net.HardwareAddr
	IPv4      *net.IPNet
	IPv6      *net.IPNet

	dbExists bool
	dbIndex  uint64
}

var _ datastore.KVObject = (*EndpointConfiguration)(nil)

// Key to be used in the KV store.
func (e *EndpointConfiguration) Key() []string {
	return []string{KVNetworkPrefix, e.ID}
}

// KeyPrefix is the immediate parent key for use in tree walks.
func (e *EndpointConfiguration) KeyPrefix() []string {
	return []string{KVNetworkPrefix}
}

// Value marshals the persistent endpoint configuration as JSON into a byte
// slice to be stored in the KV store. In case of a marshalling error, it
// returns a nil slice.
func (e *EndpointConfiguration) Value() []byte {
	b, err := json.Marshal(e)
	if err != nil {
		slog.Error("cannot marshal network configuration data",
			slog.String("network", e.ID),
			xslog.Error(err))
		return nil
	}
	return b
}

// SetValue is used by the KV store to set the object's value when loaded from
// the data store.
func (e *EndpointConfiguration) SetValue(v []byte) error {
	return json.Unmarshal(v, e)
}

// Index returns the latest DB index as seen by this configuration object.
func (e *EndpointConfiguration) Index() uint64 { return e.dbIndex }

// SetIndex allows the data store to set the latest DB Index for this
// configuration object.
func (e *EndpointConfiguration) SetIndex(i uint64) {
	e.dbExists = true
	e.dbIndex = i
}

// Exists returns true if this configuration object exists in the data store;
// otherwise false if it hasn't been stored yet. When
// [EndpointConfiguration.SetIndex] is called, the object has been stored.
func (e *EndpointConfiguration) Exists() bool { return e.dbExists }

// Skip returning true, allows a KV object avoiding getting persisted in the
// data store.
func (e *EndpointConfiguration) Skip() bool { return false }

// New returns a new zero-value configuration object. Please note that
// [KVObject's New comment] could be read as behaving somewhat clone-ish, but
// libnetwork's internal drivers, such as macvlan, show that this really is
// new-ish returning a new zero value configuration object.
//
// [KVObject's New comment]: https://pkg.go.dev/github.com/moby/moby/v2/daemon/libnetwork/datastore#KVObject
func (e *EndpointConfiguration) New() datastore.KVObject {
	return &EndpointConfiguration{}
}

// CopyTo deep copies the configuration of this object into the passed
// destination object.
func (e *EndpointConfiguration) CopyTo(other datastore.KVObject) error {
	*(other.(*EndpointConfiguration)) = *e // shallow copy
	return nil
}
