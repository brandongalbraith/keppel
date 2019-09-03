/******************************************************************************
*
*  Copyright 2018-2019 SAP SE
*
*  Licensed under the Apache License, Version 2.0 (the "License");
*  you may not use this file except in compliance with the License.
*  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
*  Unless required by applicable law or agreed to in writing, software
*  distributed under the License is distributed on an "AS IS" BASIS,
*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*  See the License for the specific language governing permissions and
*  limitations under the License.
*
******************************************************************************/

package keppelv1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sapcc/go-bits/respondwith"
	"github.com/sapcc/keppel/internal/keppel"
)

////////////////////////////////////////////////////////////////////////////////
// data types

//Account represents an account in the API.
type Account struct {
	Name         string       `json:"name"`
	AuthTenantID string       `json:"auth_tenant_id"`
	RBACPolicies []RBACPolicy `json:"rbac_policies"`
}

//RBACPolicy represents an RBAC policy in the API.
type RBACPolicy struct {
	RepositoryPattern string   `json:"match_repository,omitempty"`
	UserNamePattern   string   `json:"match_username,omitempty"`
	Permissions       []string `json:"permissions"`
}

////////////////////////////////////////////////////////////////////////////////
// data conversion/validation functions

func (a *API) renderAccount(dbAccount keppel.Account) (Account, error) {
	var dbPolicies []keppel.RBACPolicy
	_, err := a.db.Select(&dbPolicies, `SELECT * FROM rbac_policies WHERE account_name = $1`, dbAccount.Name)
	if err != nil {
		return Account{}, err
	}

	policies := make([]RBACPolicy, len(dbPolicies))
	for idx, p := range dbPolicies {
		policies[idx] = renderRBACPolicy(p)
	}

	return Account{
		Name:         dbAccount.Name,
		AuthTenantID: dbAccount.AuthTenantID,
		RBACPolicies: policies,
	}, nil
}

func renderRBACPolicy(dbPolicy keppel.RBACPolicy) RBACPolicy {
	result := RBACPolicy{
		RepositoryPattern: dbPolicy.RepositoryPattern,
		UserNamePattern:   dbPolicy.UserNamePattern,
	}
	if dbPolicy.CanPullAnonymously {
		result.Permissions = append(result.Permissions, "anonymous_pull")
	}
	if dbPolicy.CanPull {
		result.Permissions = append(result.Permissions, "pull")
	}
	if dbPolicy.CanPush {
		result.Permissions = append(result.Permissions, "push")
	}
	return result
}

func parseRBACPolicy(policy RBACPolicy) (keppel.RBACPolicy, error) {
	result := keppel.RBACPolicy{
		RepositoryPattern: policy.RepositoryPattern,
		UserNamePattern:   policy.UserNamePattern,
	}
	for _, perm := range policy.Permissions {
		switch perm {
		case "anonymous_pull":
			result.CanPullAnonymously = true
		case "pull":
			result.CanPull = true
		case "push":
			result.CanPush = true
		default:
			return result, fmt.Errorf("%q is not a valid RBAC policy permission", perm)
		}
	}

	if len(policy.Permissions) == 0 {
		return result, errors.New(`RBAC policy must grant at least one permission`)
	}
	if result.UserNamePattern == "" && result.RepositoryPattern == "" {
		return result, errors.New(`RBAC policy must have at least one "match_..." attribute`)
	}
	if result.CanPullAnonymously && result.UserNamePattern != "" {
		return result, errors.New(`RBAC policy with "anonymous_pull" may not have the "match_username" attribute`)
	}
	if result.CanPull && result.UserNamePattern == "" {
		return result, errors.New(`RBAC policy with "pull" must have the "match_username" attribute`)
	}
	if result.CanPush && !result.CanPull {
		return result, errors.New(`RBAC policy with "push" must also grant "pull"`)
	}

	for _, pattern := range []string{policy.RepositoryPattern, policy.UserNamePattern} {
		if pattern == "" {
			continue
		}
		if _, err := regexp.Compile(pattern); err != nil {
			return result, fmt.Errorf("%q is not a valid regex: %s", pattern, err.Error())
		}
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
// handlers

func (a *API) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	authz, authErr := a.authDriver.AuthenticateUserFromRequest(r)
	if respondWithAuthError(w, authErr) {
		return
	}

	var accounts []keppel.Account
	_, err := a.db.Select(&accounts, "SELECT * FROM accounts ORDER BY name")
	if respondwith.ErrorText(w, err) {
		return
	}

	//restrict accounts to those visible in the current scope
	var accountsFiltered []keppel.Account
	for _, account := range accounts {
		if authz.HasPermission(keppel.CanViewAccount, account.AuthTenantID) {
			accountsFiltered = append(accountsFiltered, account)
		}
	}
	//ensure that this serializes as a list, not as null
	if len(accountsFiltered) == 0 {
		accountsFiltered = []keppel.Account{}
	}

	//render accounts to JSON
	accountsRendered := make([]Account, len(accountsFiltered))
	for idx, account := range accountsFiltered {
		accountsRendered[idx], err = a.renderAccount(account)
		if respondwith.ErrorText(w, err) {
			return
		}
	}
	respondwith.JSON(w, http.StatusOK, map[string]interface{}{"accounts": accountsRendered})
}

func (a *API) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	authz, authErr := a.authDriver.AuthenticateUserFromRequest(r)
	if respondWithAuthError(w, authErr) {
		return
	}

	//get account from DB to find its AuthTenantID
	accountName := mux.Vars(r)["account"]
	account, err := a.db.FindAccount(accountName)
	if respondwith.ErrorText(w, err) {
		return
	}

	//perform final authorization with that AuthTenantID
	if account != nil && !authz.HasPermission(keppel.CanViewAccount, account.AuthTenantID) {
		account = nil
	}

	//this returns 404 even if the real reason is lack of authorization in order
	//to not leak information about which accounts exist for other tenants
	if account == nil {
		http.Error(w, "no such account", 404)
		return
	}

	accountRendered, err := a.renderAccount(*account)
	if respondwith.ErrorText(w, err) {
		return
	}
	respondwith.JSON(w, http.StatusOK, map[string]interface{}{"account": accountRendered})
}

func (a *API) handlePutAccount(w http.ResponseWriter, r *http.Request) {
	//decode request body
	var req struct {
		Account struct {
			AuthTenantID string       `json:"auth_tenant_id"`
			RBACPolicies []RBACPolicy `json:"rbac_policies"`
		} `json:"account"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "request body is not valid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := a.authDriver.ValidateTenantID(req.Account.AuthTenantID); err != nil {
		http.Error(w, `malformed attribute "account.auth_tenant_id" in request body: `+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	//reserve identifiers for internal pseudo-accounts
	accountName := mux.Vars(r)["account"]
	if strings.HasPrefix(accountName, "keppel-") {
		http.Error(w, `account names with the prefix "keppel-" are reserved for internal use`, http.StatusUnprocessableEntity)
		return
	}

	rbacPolicies := make([]keppel.RBACPolicy, len(req.Account.RBACPolicies))
	for idx, policy := range req.Account.RBACPolicies {
		rbacPolicies[idx], err = parseRBACPolicy(policy)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}

	accountToCreate := keppel.Account{
		Name:         accountName,
		AuthTenantID: req.Account.AuthTenantID,
	}

	//check permission to create account
	authz, authErr := a.authDriver.AuthenticateUserFromRequest(r)
	if respondWithAuthError(w, authErr) {
		return
	}
	if !authz.HasPermission(keppel.CanChangeAccount, accountToCreate.AuthTenantID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	//check if account already exists
	account, err := a.db.FindAccount(accountName)
	if respondwith.ErrorText(w, err) {
		return
	}
	if account != nil && account.AuthTenantID != req.Account.AuthTenantID {
		http.Error(w, `account name already in use by a different tenant`, http.StatusConflict)
		return
	}

	//create account if required
	if account == nil {
		tx, err := a.db.Begin()
		if respondwith.ErrorText(w, err) {
			return
		}
		defer keppel.RollbackUnlessCommitted(tx)

		account = &accountToCreate
		err = tx.Insert(account)
		if respondwith.ErrorText(w, err) {
			return
		}

		//before committing this, add the required role assignments
		err = a.authDriver.SetupAccount(*account, authz)
		if respondwith.ErrorText(w, err) {
			return
		}
		err = tx.Commit()
		if respondwith.ErrorText(w, err) {
			return
		}
	}

	err = a.putRBACPolicies(*account, rbacPolicies)
	if respondwith.ErrorText(w, err) {
		return
	}

	accountRendered, err := a.renderAccount(*account)
	if respondwith.ErrorText(w, err) {
		return
	}
	respondwith.JSON(w, http.StatusOK, map[string]interface{}{"account": accountRendered})
}

func (a *API) putRBACPolicies(account keppel.Account, policies []keppel.RBACPolicy) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer keppel.RollbackUnlessCommitted(tx)

	//delete all existing policies
	_, err = tx.Exec(`DELETE FROM rbac_policies WHERE account_name = $1`, account.Name)
	if err != nil {
		return err
	}

	//insert the requested policies
	policiesForInsert := make([]interface{}, len(policies))
	for idx, policy := range policies {
		//need to clone because `policy` gets overwritten in the next loop iteration
		cloned := policy
		cloned.AccountName = account.Name
		policiesForInsert[idx] = &cloned
	}
	err = tx.Insert(policiesForInsert...)
	if err != nil {
		return err
	}

	return tx.Commit()
}
