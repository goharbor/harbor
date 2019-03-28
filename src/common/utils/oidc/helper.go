// Copyright 2018 Project Harbor Authors
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

package oidc

import (
	"context"
	"fmt"
	gooidc "github.com/coreos/go-oidc"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"golang.org/x/oauth2"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const googleEndpoint = "https://accounts.google.com"

type providerHelper struct {
	sync.Mutex
	ep       atomic.Value
	instance atomic.Value
	setting  atomic.Value
}

func (p *providerHelper) get() (*gooidc.Provider, error) {
	if p.instance.Load() != nil {
		if p.ep.Load().(string) != p.setting.Load().(models.OIDCSetting).Endpoint {
			if err := p.create(); err != nil {
				return nil, err
			}
		}
	} else {
		p.Lock()
		defer p.Unlock()
		if p.instance.Load() == nil {
			if err := p.loadConf(); err != nil {
				return nil, err
			}
			if err := p.create(); err != nil {
				return nil, err
			}
			go func() {
				for {
					if err := p.loadConf(); err != nil {
						log.Warningf(err.Error())
					}
					time.Sleep(3 * time.Second)
				}
			}()
		}
	}

	return p.instance.Load().(*gooidc.Provider), nil

}

func (p *providerHelper) loadConf() error {
	var c *models.OIDCSetting
	c, err := config.OIDCSetting()
	if err != nil {
		return fmt.Errorf("failed to load OIDC setting: %v", err)
	}
	p.setting.Store(*c)
	return nil
}

func (p *providerHelper) create() error {
	bc := context.Background()
	s := p.setting.Load().(models.OIDCSetting)
	provider, err := gooidc.NewProvider(bc, s.Endpoint)
	if err != nil {
		return err
	}
	p.ep.Store(s.Endpoint)
	p.instance.Store(provider)
	return nil
}

var provider = &providerHelper{}

func getOauthConf() (*oauth2.Config, error) {
	p, err := provider.get()
	if err != nil {
		return nil, err
	}
	setting := provider.setting.Load().(models.OIDCSetting)
	scopes := []string{}
	for _, sc := range setting.Scope {
		if strings.HasPrefix(p.Endpoint().AuthURL, googleEndpoint) && sc == gooidc.ScopeOfflineAccess {
			log.Warningf("Dropped unsupported scope: %s ", sc)
			continue
		}
		scopes = append(scopes, sc)
	}
	return &oauth2.Config{
		ClientID:     setting.ClientID,
		ClientSecret: setting.ClientSecret,
		Scopes:       scopes,
		RedirectURL:  setting.RedirectURL,
		Endpoint:     p.Endpoint(),
	}, nil
}

// AuthCodeURL returns the URL for OIDC provider's consent page.  The state should be verified when user is redirected
// back to Harbor.
func AuthCodeURL(state string) (string, error) {
	conf, err := getOauthConf()
	if err != nil {
		log.Errorf("Failed to get OAuth configuration, error: %v", err)
		return "", err
	}
	if strings.HasPrefix(conf.Endpoint.AuthURL, googleEndpoint) {
		return conf.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
	}
	return conf.AuthCodeURL(state), nil
}
