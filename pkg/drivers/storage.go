/*******************************************************************************
*
* Copyright 2018 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/

package drivers

import (
	"errors"

	"github.com/sapcc/keppel/pkg/database"
)

//StorageDriver is the abstract interface for a multi-tenant-capable storage
//backend where the keppel-registry fleet can store images.
type StorageDriver interface {
	//ReadConfig unmarshals the configuration for this driver type into this
	//driver instance. The `unmarshal` function works exactly like in
	//UnmarshalYAML. This method shall only fail if the input data is malformed.
	//It shall not make any network requests.
	ReadConfig(unmarshal func(interface{}) error) error

	//GetEnvironment produces the environment variables (in the standard
	//"key=value" format) that need to be passed to a keppel-registry process to
	//set it up to read from/write to this storage. `tenantID` identifies the
	//tenant which controls access to this account.
	//
	//The tenant is backed by the given AuthDriver. Implementations should
	//inspect the driver to ensure that the storage backend can work with this
	//authentication method, returning ErrAuthDriverMismatch otherwise.
	GetEnvironment(account database.Account, driver AuthDriver) ([]string, error)
}

//Error types used by StorageDriver.
var (
	ErrAuthDriverMismatch = errors.New("given AuthDriver is not supported by this StorageDriver")
)