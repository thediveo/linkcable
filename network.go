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
	"sync"
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

// addNetwork adds the passed network configuration
func (d *Driver) addNetwork() error {

}
