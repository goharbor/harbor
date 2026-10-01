// Copyright Project Harbor Authors
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

package store

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config/metadata"
)

type revisionedDriver struct {
	mu       sync.Mutex
	values   map[string]any
	revision uint64
	loads    int
	saved    map[string]any
	saveErr  error
	// loadGate, when set, blocks Load until it is closed; loadEntered is closed
	// once a Load is blocked on it
	loadGate    chan struct{}
	loadEntered chan struct{}
}

func (d *revisionedDriver) Load(context.Context) (map[string]any, error) {
	d.mu.Lock()
	gate, entered := d.loadGate, d.loadEntered
	d.mu.Unlock()
	if gate != nil {
		close(entered)
		<-gate
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.loads++
	out := map[string]any{}
	for k, v := range d.values {
		out[k] = v
	}
	return out, nil
}

func (d *revisionedDriver) Save(_ context.Context, cfg map[string]any) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.saveErr != nil {
		return d.saveErr
	}
	d.saved = cfg
	return nil
}

func (d *revisionedDriver) Get(context.Context, string) (map[string]any, error) { return nil, nil }

func (d *revisionedDriver) Revision() uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.revision
}

func (d *revisionedDriver) publish(values map[string]any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.values = values
	d.revision++
}

func TestLoadSkipsUnchangedRevision(t *testing.T) {
	d := &revisionedDriver{}
	d.publish(map[string]any{common.AUTHMode: "db_auth"})
	s := NewConfigStore(d)

	require.NoError(t, s.Load(context.Background()))
	require.NoError(t, s.Load(context.Background()))
	assert.Equal(t, 1, d.loads)

	d.publish(map[string]any{common.AUTHMode: "oidc_auth"})
	require.NoError(t, s.Load(context.Background()))
	assert.Equal(t, 2, d.loads)
	v, err := s.Get(common.AUTHMode)
	require.NoError(t, err)
	assert.Equal(t, "oidc_auth", v.GetString())
}

func TestUpdateIsNotPublishedWhenSaveFails(t *testing.T) {
	d := &revisionedDriver{}
	d.publish(map[string]any{common.AUTHMode: "db_auth"})
	s := NewConfigStore(d)
	require.NoError(t, s.Load(context.Background()))

	d.saveErr = errors.New("save failed")
	require.Error(t, s.Update(context.Background(), map[string]any{common.AUTHMode: "ldap_auth"}))
	v, _ := s.Get(common.AUTHMode)
	assert.Equal(t, "db_auth", v.GetString())
}

func TestUpdateIsPublishedWithoutTransaction(t *testing.T) {
	d := &revisionedDriver{}
	d.publish(map[string]any{common.AUTHMode: "db_auth"})
	s := NewConfigStore(d)
	require.NoError(t, s.Load(context.Background()))

	require.NoError(t, s.Update(context.Background(), map[string]any{common.AUTHMode: "ldap_auth"}))
	assert.Equal(t, "ldap_auth", d.saved[common.AUTHMode])
	// a Load before the driver announces the change keeps the committed local value
	require.NoError(t, s.Load(context.Background()))
	v, _ := s.Get(common.AUTHMode)
	assert.Equal(t, "ldap_auth", v.GetString())
}

// the SyncQuota pattern
func TestFailedSaveIsUndoneByNextLoad(t *testing.T) {
	d := &revisionedDriver{}
	d.publish(map[string]any{common.ReadOnly: "false"})
	s := NewConfigStore(d)
	require.NoError(t, s.Load(context.Background()))

	require.NoError(t, s.Set(common.ReadOnly, mustValue(t, common.ReadOnly, "true")))
	d.saveErr = errors.New("save failed")
	require.Error(t, s.Save(context.Background()))
	require.NoError(t, s.Load(context.Background()))
	v, _ := s.Get(common.ReadOnly)
	assert.False(t, v.GetBool())
}

func TestOlderLoadNeverOverwritesNewer(t *testing.T) {
	d := &revisionedDriver{}
	d.publish(map[string]any{common.AUTHMode: "db_auth"})
	s := NewConfigStore(d)
	require.NoError(t, s.Load(context.Background()))

	// the slow Load reads revision 2 and blocks inside the driver
	d.publish(map[string]any{common.AUTHMode: "ldap_auth"})
	gate, entered := blockLoads(d)
	slow := make(chan error, 1)
	go func() { slow <- s.Load(context.Background()) }()
	<-entered

	// revision 3 is merged by a fast Load while the slow one waits
	d.mu.Lock()
	d.loadGate, d.loadEntered = nil, nil
	d.values = map[string]any{common.AUTHMode: "oidc_auth"}
	d.revision++
	d.mu.Unlock()
	require.NoError(t, s.Load(context.Background()))
	require.NoError(t, s.Save(context.Background()))

	d.mu.Lock()
	d.values = map[string]any{common.AUTHMode: "ldap_auth"} // what the slow Load will read
	d.mu.Unlock()
	close(gate)
	require.NoError(t, <-slow)
	v, _ := s.Get(common.AUTHMode)
	assert.Equal(t, "oidc_auth", v.GetString())
}

func TestZeroRevisionAlwaysLoads(t *testing.T) {
	d := &revisionedDriver{values: map[string]any{common.AUTHMode: "db_auth"}}
	s := NewConfigStore(d)
	require.NoError(t, s.Load(context.Background()))
	require.NoError(t, s.Load(context.Background()))
	assert.Equal(t, 2, d.loads)
}

func TestLoadSwapsValuesAtomically(t *testing.T) {
	gen := func(i int) map[string]any {
		n := strconv.Itoa(i)
		return map[string]any{common.OIDCName: "name-" + n, common.OIDCEndpoint: "https://idp-" + n}
	}
	d := &revisionedDriver{}
	d.publish(gen(0))
	s := NewConfigStore(d)
	require.NoError(t, s.Load(context.Background()))

	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i < 2000; i++ {
			d.publish(gen(i))
			_ = s.Load(context.Background())
		}
		close(stop)
	}()
	for {
		select {
		case <-stop:
			wg.Wait()
			return
		default:
		}
		m := s.values()
		name, endpoint := m[common.OIDCName], m[common.OIDCEndpoint]
		require.Equal(t, name.Value[len("name-"):], endpoint.Value[len("https://idp-"):])
	}
}

func TestSetOnStoreWithoutValues(t *testing.T) {
	s := &ConfigStore{}
	_, err := s.Get(common.AUTHMode)
	assert.Error(t, err)
	require.NoError(t, s.Set(common.AUTHMode, mustValue(t, common.AUTHMode, "db_auth")))
	v, err := s.Get(common.AUTHMode)
	require.NoError(t, err)
	assert.Equal(t, "db_auth", v.GetString())
}

func mustValue(t *testing.T, key, value string) metadata.ConfigureValue {
	t.Helper()
	v, err := metadata.NewCfgValue(key, value)
	require.NoError(t, err)
	return *v
}

func blockLoads(d *revisionedDriver) (gate, entered chan struct{}) {
	gate, entered = make(chan struct{}), make(chan struct{})
	d.mu.Lock()
	d.loadGate, d.loadEntered = gate, entered
	d.mu.Unlock()
	return gate, entered
}

func TestInFlightLoadDoesNotOverwriteUpdate(t *testing.T) {
	d := &revisionedDriver{}
	d.publish(map[string]any{common.AUTHMode: "db_auth"})
	s := NewConfigStore(d)
	require.NoError(t, s.Load(context.Background()))

	// force the next Load to read the driver, then hold it there
	require.NoError(t, s.Save(context.Background()))
	gate, entered := blockLoads(d)
	slow := make(chan error, 1)
	go func() { slow <- s.Load(context.Background()) }()
	<-entered

	d.mu.Lock()
	d.loadGate, d.loadEntered = nil, nil
	d.mu.Unlock()
	require.NoError(t, s.Update(context.Background(), map[string]any{common.AUTHMode: "ldap_auth"}))
	close(gate)
	require.NoError(t, <-slow)
	v, _ := s.Get(common.AUTHMode)
	assert.Equal(t, "ldap_auth", v.GetString())
}
