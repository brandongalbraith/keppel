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

package keppel

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	gorp "gopkg.in/gorp.v2"
)

//Account contains a record from the `accounts` table.
type Account struct {
	Name         string `db:"name"`
	AuthTenantID string `db:"auth_tenant_id"`
}

//SwiftContainerName returns the name of the Swift container backing this
//Keppel account.
func (a Account) SwiftContainerName() string {
	return "keppel-" + a.Name
}

//PostgresDatabaseName returns the name of the Postgres database which contains this
//Keppel account's metadata.
func (a Account) PostgresDatabaseName() string {
	return "keppel_" + strings.Replace(a.Name, "-", "_", -1)
}

//FindAccount works similar to db.SelectOne(), but returns nil instead of
//sql.ErrNoRows if no account exists with this name.
func (db *DB) FindAccount(name string) (*Account, error) {
	var account Account
	err := db.SelectOne(&account,
		"SELECT * FROM accounts WHERE name = $1", name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &account, err
}

//AllAccounts implements the DBAccessForOrchestrationDriver interface.
func (db *DB) AllAccounts() ([]Account, error) {
	var accounts []Account
	_, err := db.Select(&accounts, `SELECT * FROM accounts`)
	if err != nil {
		accounts = nil
	}
	return accounts, err
}

//RBACPolicy contains a record from the `rbac_policies` table.
type RBACPolicy struct {
	AccountName        string `db:"account_name"`
	RepositoryPattern  string `db:"match_repository"`
	UserNamePattern    string `db:"match_username"`
	CanPullAnonymously bool   `db:"can_anon_pull"`
	CanPull            bool   `db:"can_pull"`
	CanPush            bool   `db:"can_push"`
}

//Matches evaluates the regexes in this policy.
func (r RBACPolicy) Matches(repoName, userName string) bool {
	if r.RepositoryPattern != "" {
		rx, err := regexp.Compile(fmt.Sprintf(`^%s/%s$`,
			regexp.QuoteMeta(r.AccountName),
			r.RepositoryPattern,
		))
		if err != nil || !rx.MatchString(repoName) {
			return false
		}
	}

	if r.UserNamePattern != "" {
		rx, err := regexp.Compile(fmt.Sprintf(`^%s$`, r.UserNamePattern))
		if err != nil || !rx.MatchString(userName) {
			return false
		}
	}

	return true
}

func initModels(db *gorp.DbMap) {
	db.AddTableWithName(Account{}, "accounts").SetKeys(false, "name")
	db.AddTableWithName(RBACPolicy{}, "rbac_policies").SetKeys(false, "account_name", "match_repository", "match_username")
}
