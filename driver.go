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
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/docker/go-plugins-helpers/network"
	"github.com/moby/moby/v2/daemon/libnetwork/datastore"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"

	"github.com/thediveo/linkcable/privy"
)

// Driver drives netkit-powered virtual point-to-point links, similar to what
// veth does, but with container-facing link ends with independent lifecycles.
type Driver struct {
	name  string
	store *datastore.Store

	// fd referencing a privy network namespace for primary netkit interfaces;
	// zero value means unset (we leave fd 0 to whatever is attached there at
	// startup).
	privynetnsfd int
	// netlink handle/connection to the privy network namespace to carry out
	// operations in that network namespace. We will be placing the primary
	// netkit interfaces there so they do their packet forwarding in splendid
	// isolation.
	privyhdl *netlink.Handle
	// netlink handle/connection to the "host", or rather: current, network
	// namespace to carry out operations in that network namespace. We will be
	// placing the peer netkit interfaces there so that libnetwork can handle
	// them; yeah, this is an unfortunate architectural shortcoming of
	// libnetwork.
	hosthdl *netlink.Handle

	m        sync.RWMutex
	networks NetworksByID
}

var _ network.Driver = (*Driver)(nil)

func NewDriver(name string, tabularasa bool) (*Driver, error) {
	dbPath := "/var/lib/linkcable/linkcable"
	privynetnsPath := "/run/linkcable/privynetns"
	if name != "linkcable" {
		dbPath += "-" + name
		privynetnsPath += "-" + name
	}
	slog.Info("persistent driver configuration", slog.String("path", dbPath))
	slog.Info("privy network namespace reference", slog.String("path", privynetnsPath))

	store, err := datastore.New(dbPath, "linkcable")
	if err != nil {
		return nil, fmt.Errorf("cannot create or reuse boltdb store %q, reason: %w",
			filepath.Join(dbPath, "local-kv.db"), // mirrors datastore.New behavior
			err)
	}
	d := &Driver{
		name:  name,
		store: store,
	}

	d.privynetnsfd, err = privy.BindmountedNetns(privynetnsPath)
	if err != nil {
		_ = d.Close()
		return nil, err
	}

	d.privyhdl, err = netlink.NewHandleAt(netns.NsHandle(d.privynetnsfd))
	if err != nil {
		_ = d.Close()
		return nil, err
	}

	d.hosthdl, err = netlink.NewHandle()
	if err != nil {
		_ = d.Close()
		return nil, err
	}

	if err := d.restoreNetworks(tabularasa); err != nil {
		_ = d.Close()
		return nil, err
	}
	return d, nil
}

func (d *Driver) Close() error {
	if d.hosthdl != nil {
		_ = d.hosthdl.Close()
	}
	if d.privyhdl != nil {
		_ = d.privyhdl.Close()
	}
	if d.privynetnsfd > 0 {
		_ = unix.Close(d.privynetnsfd)
	}
	if d.store != nil {
		d.store.Close()
	}
	return nil
}

func (*Driver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	return &network.CapabilitiesResponse{
		Scope:             network.LocalScope,
		ConnectivityScope: network.LocalScope,
	}, nil
}

func (*Driver) CreateNetwork(*network.CreateNetworkRequest) error {

}

func (*Driver) DeleteNetwork(*network.DeleteNetworkRequest) error {
	return errors.New("not implemented")
}

func (*Driver) AllocateNetwork(*network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	return nil, errors.New("not implemented")
}
func (*Driver) FreeNetwork(*network.FreeNetworkRequest) error {
	return errors.New("not implemented")
}
func (*Driver) CreateEndpoint(*network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	return nil, errors.New("not implemented")
}
func (*Driver) DeleteEndpoint(*network.DeleteEndpointRequest) error {
	return errors.New("not implemented")
}
func (*Driver) EndpointInfo(*network.InfoRequest) (*network.InfoResponse, error) {
	return nil, errors.New("not implemented")
}
func (*Driver) Join(*network.JoinRequest) (*network.JoinResponse, error) {
	return nil, errors.New("not implemented")
}
func (*Driver) Leave(*network.LeaveRequest) error {
	return errors.New("not implemented")
}
func (*Driver) DiscoverNew(*network.DiscoveryNotification) error {
	return errors.New("not implemented")
}
func (*Driver) DiscoverDelete(*network.DiscoveryNotification) error {
	return errors.New("not implemented")
}
func (*Driver) ProgramExternalConnectivity(*network.ProgramExternalConnectivityRequest) error {
	return errors.New("not implemented")
}
func (*Driver) RevokeExternalConnectivity(*network.RevokeExternalConnectivityRequest) error {
	return errors.New("not implemented")
}
