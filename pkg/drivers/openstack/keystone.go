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

//Package openstack contains:
//
//- the AuthDriver "keystone": Keppel tenants are Keystone projects. Incoming HTTP requests are authenticated by reading a Keystone token from the X-Auth-Token request header.
//
//- the StorageDriver "swift": Data for a Keppel account is stored in the Swift container "keppel-<accountname>" in the tenant's Swift account.
package openstack

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/roles"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	"github.com/sapcc/go-bits/gopherpolicy"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/keppel/pkg/database"
	"github.com/sapcc/keppel/pkg/drivers"
)

type keystoneDriver struct {
	//configuration
	ServiceUser struct {
		AuthURL           string `yaml:"auth_url"`
		UserName          string `yaml:"user_name"`
		UserDomainName    string `yaml:"user_domain_name"`
		ProjectName       string `yaml:"project_name"`
		ProjectDomainName string `yaml:"project_domain_name"`
		Password          string `yaml:"password"`
	} `yaml:"service_user"`
	LocalRoleName  string `yaml:"local_role"`
	PolicyFilePath string `yaml:"policy_path"`
	//TODO remove when https://github.com/gophercloud/gophercloud/issues/1141 is accepted
	UserID string `yaml:"user_id"`

	Client         *gophercloud.ProviderClient  `yaml:"-"`
	IdentityV3     *gophercloud.ServiceClient   `yaml:"-"`
	TokenValidator *gopherpolicy.TokenValidator `yaml:"-"`
	LocalRoleID    string                       `yaml:"-"`
}

func init() {
	drivers.RegisterAuthDriver("keystone", func() drivers.AuthDriver {
		return &keystoneDriver{}
	})
}

//ReadConfig implements the drivers.AuthDriver interface.
func (d *keystoneDriver) ReadConfig(unmarshal func(interface{}) error) error {
	err := unmarshal(d)
	if err != nil {
		return err
	}
	if d.ServiceUser.AuthURL == "" {
		return errors.New("missing auth.service_user.auth_url")
	}
	if d.ServiceUser.UserName == "" {
		return errors.New("missing auth.service_user.user_name")
	}
	if d.ServiceUser.UserDomainName == "" {
		return errors.New("missing auth.service_user.user_domain_name")
	}
	if d.ServiceUser.Password == "" {
		return errors.New("missing auth.service_user.password")
	}
	if d.ServiceUser.ProjectName == "" {
		return errors.New("missing auth.service_user.project_name")
	}
	if d.ServiceUser.ProjectDomainName == "" {
		return errors.New("missing auth.service_user.project_domain_name")
	}
	if d.LocalRoleName == "" {
		return errors.New("missing auth.local_role")
	}
	if d.PolicyFilePath == "" {
		return errors.New("missing auth.policy_path")
	}
	if d.UserID == "" {
		return errors.New("missing auth.user_id")
	}
	return nil
}

//Connect implements the drivers.AuthDriver interface.
func (d *keystoneDriver) Connect() error {
	var err error
	d.Client, err = openstack.NewClient(d.ServiceUser.AuthURL)
	if err != nil {
		logg.Fatal("cannot initialize OpenStack client: %v", err)
	}

	//use http.DefaultClient, esp. to pick up the KEPPEL_INSECURE flag
	d.Client.HTTPClient = *http.DefaultClient

	err = openstack.Authenticate(d.Client, gophercloud.AuthOptions{
		IdentityEndpoint: d.ServiceUser.AuthURL,
		AllowReauth:      true,
		Username:         d.ServiceUser.UserName,
		DomainName:       d.ServiceUser.UserDomainName,
		Password:         d.ServiceUser.Password,
		Scope: &gophercloud.AuthScope{
			ProjectName: d.ServiceUser.ProjectName,
			DomainName:  d.ServiceUser.ProjectDomainName,
		},
	})
	if err != nil {
		return fmt.Errorf("cannot fetch initial Keystone token: %v", err)
	}

	d.IdentityV3, err = openstack.NewIdentityV3(d.Client, gophercloud.EndpointOpts{})
	if err != nil {
		return fmt.Errorf("cannot find Identity v3 API in Keystone catalog: %s", err.Error())
	}

	d.TokenValidator = &gopherpolicy.TokenValidator{
		IdentityV3: d.IdentityV3,
	}
	err = d.TokenValidator.LoadPolicyFile(d.PolicyFilePath)
	if err != nil {
		return err
	}

	localRole, err := getRoleByName(d.IdentityV3, d.LocalRoleName)
	if err != nil {
		return fmt.Errorf("cannot find Keystone role '%s': %s", d.LocalRoleName, err.Error())
	}
	d.LocalRoleID = localRole.ID

	return nil
}

func getRoleByName(identityV3 *gophercloud.ServiceClient, name string) (roles.Role, error) {
	page, err := roles.List(identityV3, roles.ListOpts{Name: name}).AllPages()
	if err != nil {
		return roles.Role{}, err
	}
	list, err := roles.ExtractRoles(page)
	if err != nil {
		return roles.Role{}, err
	}
	if len(list) == 0 {
		return roles.Role{}, errors.New("no such role")
	}
	return list[0], nil
}

//SetupAccount implements the drivers.AuthDriver interface.
func (d *keystoneDriver) SetupAccount(account database.Account, authorization drivers.Authorization) error {
	requesterToken := authorization.(keystoneAuthorization).t //is a *gopherpolicy.Token
	client, err := openstack.NewIdentityV3(
		requesterToken.ProviderClient, gophercloud.EndpointOpts{})
	if err != nil {
		return err
	}
	result := roles.Assign(client, d.LocalRoleID, roles.AssignOpts{
		UserID:    d.UserID,
		ProjectID: account.ProjectUUID,
	})
	return result.Err
}

//AuthenticateUser implements the drivers.AuthDriver interface.
func (d *keystoneDriver) AuthenticateUser(userName, password string) (drivers.Authorization, error) {
	return d.AuthenticateUserInTenant(userName, password, "")
}

//AuthenticateUserInTenant implements the drivers.AuthDriver interface.
func (d *keystoneDriver) AuthenticateUserInTenant(userName, password, tenantID string) (drivers.Authorization, error) {
	usernameFields := strings.SplitN(userName, "@", 2)
	if len(usernameFields) != 2 {
		logg.Info(`invalid username in Authorization header (expected "user@domain" format)`)
		return nil, drivers.ErrUnauthorized
	}

	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: d.IdentityV3.Endpoint,
		Username:         usernameFields[0],
		DomainName:       usernameFields[1],
		Password:         password,
	}
	if tenantID != "" {
		authOpts.Scope = &gophercloud.AuthScope{ProjectID: tenantID}
	}

	result := tokens.Create(d.IdentityV3, &authOpts)
	t := d.TokenValidator.TokenFromGophercloudResult(result)
	if t.Err != nil {
		logg.Info("failed to get token for user %q in project %q: %s", userName, tenantID, t.Err.Error())
		return nil, drivers.ErrUnauthorized
	}
	if !t.Check("account:list") {
		return nil, drivers.ErrForbidden
	}
	return &keystoneAuthorization{t}, nil
}

//AuthenticateUserFromRequest implements the drivers.AuthDriver interface.
func (d *keystoneDriver) AuthenticateUserFromRequest(r *http.Request) (drivers.Authorization, error) {
	t := d.TokenValidator.CheckToken(r)
	if t.Err != nil {
		logg.Info("X-Auth-Token validation failed: " + t.Err.Error())
		return nil, drivers.ErrUnauthorized
	}

	t.Context.Logger = logg.Debug
	//token.Context.Request = mux.Vars(r) //not used at the moment

	if !t.Check("account:list") {
		return nil, drivers.ErrForbidden
	}
	return keystoneAuthorization{t}, nil
}

type keystoneAuthorization struct {
	t *gopherpolicy.Token
}

var ruleForPerm = map[drivers.Permission]string{
	drivers.CanViewAccount:     "account:show",
	drivers.CanPullFromAccount: "account:pull",
	drivers.CanPushToAccount:   "account:push",
	drivers.CanChangeAccount:   "account:edit",
}

//HasPermission implements the drivers.Authorization interface.
func (a keystoneAuthorization) HasPermission(perm drivers.Permission, tenantID string) bool {
	a.t.Context.Request["account_project_id"] = tenantID
	return a.t.Check(ruleForPerm[perm])
}