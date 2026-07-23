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

	"github.com/docker/go-plugins-helpers/network"
	"github.com/moby/moby/v2/daemon/libnetwork/datastore"
)

const KVStorageFile = "linkcable-kv.db"

// Driver drives netkit-powered virtual point-to-point links, similar to what
// veth does, but with container-facing link ends with independent lifecycles.
type Driver struct {
	name  string
	store *datastore.Store
}

var _ network.Driver = (*Driver)(nil)

func NewDriver(name string) (*Driver, error) {
	d := &Driver{
		name: name,
	}
	return d, nil
}

func (*Driver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	return &network.CapabilitiesResponse{
		Scope:             network.LocalScope,
		ConnectivityScope: network.LocalScope,
	}, nil
}
func (*Driver) CreateNetwork(*network.CreateNetworkRequest) error {
	return errors.New("not implemented")
}
func (*Driver) AllocateNetwork(*network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	return nil, errors.New("not implemented")
}
func (*Driver) DeleteNetwork(*network.DeleteNetworkRequest) error {
	return errors.New("not implemented")
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
