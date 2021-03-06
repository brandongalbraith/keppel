/******************************************************************************
*
*  Copyright 2019 SAP SE
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

package registryv2

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/sre"
	"github.com/sapcc/keppel/internal/api"
	"github.com/sapcc/keppel/internal/keppel"
	"github.com/sapcc/keppel/internal/processor"
)

//This implements the GET/HEAD /v2/<account>/<repository>/blobs/<digest> endpoint.
func (a *API) handleGetOrHeadBlob(w http.ResponseWriter, r *http.Request) {
	sre.IdentifyEndpoint(r, "/v2/:account/:repo/blobs/:digest")
	account, repo, token := a.checkAccountAccess(w, r, failIfRepoMissing, a.handleGetOrHeadBlobAnycast)
	if account == nil {
		return
	}
	if !a.checkRateLimit(w, r, *account, token, keppel.BlobPullAction, 1) {
		return
	}

	blobDigest, err := digest.Parse(mux.Vars(r)["digest"])
	if err != nil {
		keppel.ErrDigestInvalid.With(err.Error()).WriteAsRegistryV2ResponseTo(w, r)
		return
	}

	//locate this blob from the DB
	blob, err := keppel.FindBlobByRepository(a.db, blobDigest, *repo, *account)
	if err == sql.ErrNoRows {
		keppel.ErrBlobUnknown.With("blob does not exist in this repository").WriteAsRegistryV2ResponseTo(w, r)
		return
	}
	if respondWithError(w, r, err) {
		return
	}

	//if this blob has not been replicated...
	if blob.StorageID == "" {
		if account.UpstreamPeerHostName == "" && account.ExternalPeerURL == "" {
			//defense in depth: unbacked blobs should not exist in non-replica accounts
			keppel.ErrBlobUnknown.With("blob does not exist in this repository").WriteAsRegistryV2ResponseTo(w, r)
			return
		}

		//...answer HEAD requests with the metadata that we obtained when replicating the manifest...
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", strconv.FormatUint(blob.SizeBytes, 10))
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Docker-Content-Digest", blob.Digest)
			w.WriteHeader(http.StatusOK)
			return
		}

		//...and answer GET requests by replicating the blob contents
		responseWasWritten, err := a.processor().ReplicateBlob(*blob, *account, *repo, w)

		if err != nil {
			if responseWasWritten {
				//we cannot write to `w` if br.Execute() wrote a response there already
				logg.Error("while trying to replicate blob %s in %s/%s: %s",
					blob.Digest, account.Name, repo.Name, err.Error())
			} else if err == processor.ErrConcurrentReplication {
				//special handling for GET during ongoing replication (429 Too Many
				//Requests is not a perfect match, but it's my best guess for getting
				//clients to automatically retry the request after a few seconds)
				w.Header().Set("Retry-After", "10")
				msg := "currently replicating on a different worker, please retry in a few seconds"
				keppel.ErrTooManyRequests.With(msg).WriteAsRegistryV2ResponseTo(w, r)
			} else {
				respondWithError(w, r, err)
			}
		} else if !responseWasWritten {
			respondWithError(w, r, errors.New("blob replication yielded neither blob contents nor an error"))
		}

		return
	}

	//if a peer reverse-proxied to us to fulfil an anycast request, enforce the anycast rate limits
	isAnycast := r.Header.Get("X-Keppel-Forwarded-By") != ""
	if isAnycast {
		//AnycastBlobBytePullAction is only relevant for GET requests since it
		//limits the size of the response body (which is empty for HEAD)
		if r.Method == "GET" {
			if !a.checkRateLimit(w, r, *account, token, keppel.AnycastBlobBytePullAction, blob.SizeBytes) {
				return
			}
		}
	}

	//before we branch into different code paths, count the pull
	if r.Method == "GET" {
		l := prometheus.Labels{"account": account.Name, "auth_tenant_id": account.AuthTenantID, "method": "registry-api"}
		if strings.HasPrefix(token.UserName, "replication@") {
			l["method"] = "replication"
		} else if isAnycast {
			l["method"] = "registry-api+anycast"
		}
		api.BlobsPulledCounter.With(l).Inc()
		api.BlobBytesPulledCounter.With(l).Add(float64(blob.SizeBytes))
	}

	//prefer redirecting the client to a storage URL if the storage driver can give us one
	url, err := a.sd.URLForBlob(*account, blob.StorageID)
	if err == nil {
		w.Header().Set("Docker-Content-Digest", blob.Digest)
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != keppel.ErrCannotGenerateURL {
		respondWithError(w, r, err)
		return
	}

	//return the blob contents to the client directly (NOTE: this code path is
	//rather lazy and esp. does not support range requests because it is only
	//used by unit tests anyway; all production-grade storage drivers have a
	//functional URLForBlob implementation)
	reader, lengthBytes, err := a.sd.ReadBlob(*account, blob.StorageID)
	if respondWithError(w, r, err) {
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Length", strconv.FormatUint(lengthBytes, 10))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Docker-Content-Digest", blob.Digest)
	w.WriteHeader(http.StatusOK)
	if r.Method != "HEAD" {
		_, err = io.Copy(w, reader)
		if err != nil {
			logg.Error("unexpected error from io.Copy() while sending blob to client: %s", err.Error())
		}
	}
}

func (a *API) handleGetOrHeadBlobAnycast(w http.ResponseWriter, r *http.Request, info anycastRequestInfo) {
	//NOTE: Rate limits are enforced by the peer that we reverse-proxy to, not by
	//us. We couldn't enforce them anyway because we don't have this account.
	err := a.cfg.ReverseProxyAnycastRequestToPeer(w, r, info.PrimaryHostName)
	if respondWithError(w, r, err) {
		return
	}
	api.BlobsPulledCounter.With(info.AsPrometheusLabels()).Inc()
}

//This implements the DELETE /v2/<account>/<repository>/blobs/<digest> endpoint.
func (a *API) handleDeleteBlob(w http.ResponseWriter, r *http.Request) {
	sre.IdentifyEndpoint(r, "/v2/:account/:repo/blobs/:digest")
	account, repo, _ := a.checkAccountAccess(w, r, failIfRepoMissing, nil)
	if account == nil {
		return
	}

	blobDigest, err := digest.Parse(mux.Vars(r)["digest"])
	if err != nil {
		keppel.ErrDigestInvalid.With(err.Error()).WriteAsRegistryV2ResponseTo(w, r)
		return
	}

	blob, err := keppel.FindBlobByRepository(a.db, blobDigest, *repo, *account)
	if err == sql.ErrNoRows {
		keppel.ErrBlobUnknown.With("blob does not exist in this repository").WriteAsRegistryV2ResponseTo(w, r)
		return
	}
	if respondWithError(w, r, err) {
		return
	}

	//can only delete blob mount if it's not used by any manifests
	refCount, err := a.db.SelectInt(`SELECT COUNT(*) FROM manifest_blob_refs WHERE blob_id = $1 AND repo_id = $2`, blob.ID, repo.ID)
	if respondWithError(w, r, err) {
		return
	}
	if refCount > 0 {
		keppel.ErrUnsupported.
			With("blob %s cannot be deleted while it is referenced by manifests in this repo", blob.Digest).
			WithStatus(http.StatusMethodNotAllowed).
			WriteAsRegistryV2ResponseTo(w, r)
		return
	}

	//unmount the blob from this particular repo (if it is mounted in other
	//repos, it will still be accessible there; otherwise keppel-janitor will
	//clean it up soon)
	_, err = a.db.Exec(`DELETE FROM blob_mounts WHERE blob_id = $1 AND repo_id = $2`, blob.ID, repo.ID)
	if respondWithError(w, r, err) {
		return
	}

	w.Header().Set("Content-Length", "0")
	w.Header().Set("Docker-Content-Digest", blob.Digest)
	w.WriteHeader(http.StatusAccepted)
}
