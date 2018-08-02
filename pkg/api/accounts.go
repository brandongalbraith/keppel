/******************************************************************************
*
*  Copyright 2018 SAP SE
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

package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/respondwith"
	"github.com/sapcc/keppel/pkg/database"
	"github.com/sapcc/keppel/pkg/keppel"
	"github.com/sapcc/keppel/pkg/openstack"
)

func checkTokenOrSend401(w http.ResponseWriter, r *http.Request) openstack.AccessLevel {
	access, err := keppel.State.ServiceUser.GetAccessLevelForRequest(r)
	if err != nil || access == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}
	return access
}

func (api *KeppelV1) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	access := checkTokenOrSend401(w, r)
	if access == nil {
		return
	}
	if !access.CanViewAccounts() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var accounts []database.Account
	_, err := keppel.State.DB.Select(&accounts, "SELECT * FROM accounts ORDER BY name")
	if respondwith.ErrorText(w, err) {
		return
	}

	//restrict accounts to those visible in the current scope
	var accountsFiltered []database.Account
	for _, account := range accounts {
		if access.CanViewAccount(account) {
			accountsFiltered = append(accountsFiltered, account)
		}
	}
	//ensure that this serializes as a list, not as null
	if len(accountsFiltered) == 0 {
		accountsFiltered = []database.Account{}
	}

	respondwith.JSON(w, http.StatusOK, map[string]interface{}{"accounts": accountsFiltered})
}

func (api *KeppelV1) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	access := checkTokenOrSend401(w, r)
	if access == nil {
		return
	}

	//first very permissive check: can this user GET any accounts AT ALL?
	if !access.CanViewAccounts() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	//get account from DB to find its project ID
	accountName := mux.Vars(r)["account"]
	account, err := keppel.State.DB.FindAccount(accountName)
	if respondwith.ErrorText(w, err) {
		return
	}

	//perform final authorization with that project ID
	if account != nil && !access.CanViewAccount(*account) {
		account = nil
	}

	if account == nil {
		http.Error(w, "no such account", 404)
		return
	}

	respondwith.JSON(w, http.StatusOK, map[string]interface{}{"account": account})
}

func (api *KeppelV1) handlePutAccount(w http.ResponseWriter, r *http.Request) {
	//decode request body
	var req struct {
		Account struct {
			ProjectUUID string `json:"project_id"`
		} `json:"account"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "request body is not valid JSON: "+err.Error(), 400)
		return
	}
	if req.Account.ProjectUUID == "" {
		http.Error(w, `missing attribute "account.project_id" in request body`, 400)
		return
	}

	//reserve identifiers for internal pseudo-accounts
	accountName := mux.Vars(r)["account"]
	if strings.HasPrefix(accountName, "keppel-") {
		http.Error(w, `account names with the prefix "keppel-" are reserved for internal use`, 400)
		return
	}

	accountToCreate := database.Account{
		Name:        accountName,
		ProjectUUID: req.Account.ProjectUUID,
	}

	//check permission to create account
	access := checkTokenOrSend401(w, r)
	if access == nil {
		return
	}
	if !access.CanChangeAccount(accountToCreate) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	//check if account already exists
	account, err := keppel.State.DB.FindAccount(accountName)
	if respondwith.ErrorText(w, err) {
		return
	}
	if account != nil && account.ProjectUUID != req.Account.ProjectUUID {
		http.Error(w, `missing attribute "account.project_id" in request body`, http.StatusConflict)
		return
	}

	//create account if required
	if account == nil {
		tx, err := keppel.State.DB.Begin()
		if respondwith.ErrorText(w, err) {
			return
		}
		defer database.RollbackUnlessCommitted(tx)

		account = &accountToCreate
		err = tx.Insert(account)
		if respondwith.ErrorText(w, err) {
			return
		}

		//before committing this, add the required role assignments
		err = keppel.State.ServiceUser.AddLocalRole(req.Account.ProjectUUID, access)
		if respondwith.ErrorText(w, err) {
			return
		}
		err = tx.Commit()
		if respondwith.ErrorText(w, err) {
			return
		}
	}

	//ensure that keppel-registry is running (TODO remove, only used for testing)
	logg.Info("keppel-registry for account %s is running on %s",
		account.Name, api.orch.GetHostPortForAccount(*account))

	respondwith.JSON(w, http.StatusOK, map[string]interface{}{"account": account})
}