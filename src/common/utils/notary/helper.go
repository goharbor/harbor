// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package notary

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/docker/notary"
	"github.com/docker/notary/client"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf/data"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"

	"github.com/opencontainers/go-digest"
)

var (
	notaryCachePath = "/root/notary"
	trustPin        trustpinning.TrustPinConfig
	mockRetriever   notary.PassRetriever
)

// Target represents the json object of a target of a docker image in notary.
// The struct will be used when repository is know so it won'g contain the name of a repository.
type Target struct {
	Tag    string      `json:"tag"`
	Hashes data.Hashes `json:"hashes"`
	//TODO: update fields as needed.
}

func init() {
	mockRetriever = func(keyName, alias string, createNew bool, attempts int) (passphrase string, giveup bool, err error) {
		passphrase = "hardcode"
		giveup = false
		err = nil
		return
	}
	trustPin = trustpinning.TrustPinConfig{}
}

// GetTargets is a help function called by API to fetch signature information of a given repository.
// Per docker's convention the repository should contain the information of endpoint, i.e. it should look
// like "10.117.4.117/library/ubuntu", instead of "library/ubuntu" (fqRepo for fully-qualified repo)
func GetTargets(notaryEndpoint string, username string, fqRepo string) ([]Target, error) {
	res := []Target{}
	authorizer := auth.NewNotaryUsernameTokenAuthorizer(username, "repository", fqRepo, "pull")
	store, err := auth.NewAuthorizerStore(strings.Split(notaryEndpoint, "//")[1], true, authorizer)
	if err != nil {
		return res, err
	}
	tr := registry.NewTransport(registry.GetHTTPTransport(true), store)
	gun := data.GUN(fqRepo)
	notaryRepo, err := client.NewFileCachedNotaryRepository(notaryCachePath, gun, notaryEndpoint, tr, mockRetriever, trustPin)
	if err != nil {
		return res, err
	}
	targets, err := notaryRepo.ListTargets(data.CanonicalTargetsRole)
	if _, ok := err.(client.ErrRepositoryNotExist); ok {
		log.Errorf("Repository not exist, repo: %s, error: %v, returning empty signature", fqRepo, err)
		return res, nil
	} else if err != nil {
		return res, err
	}
	//Remove root.json such that when remote repository is removed the local cache can't be reused.
	rootJSON := path.Join(notaryCachePath, "tuf", fqRepo, "metadata/root.json")
	rmErr := os.Remove(rootJSON)
	if rmErr != nil {
		log.Warningf("Failed to clear cached root.json: %s, error: %v, when repo is removed from notary the signature status maybe incorrect", rootJSON, rmErr)
	}
	for _, t := range targets {
		res = append(res, Target{t.Name, t.Hashes})
	}
	return res, nil
}

// DigestFromTarget get a target and return the value of digest, in accordance to Docker-Content-Digest
func DigestFromTarget(t Target) (string, error) {
	sha, ok := t.Hashes["sha256"]
	if !ok {
		return "", fmt.Errorf("no valid hash, expecting sha256")
	}
	return digest.NewDigestFromHex("sha256", hex.EncodeToString(sha)).String(), nil
}
