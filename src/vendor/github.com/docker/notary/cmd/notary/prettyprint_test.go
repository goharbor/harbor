package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/docker/notary/client"
	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

// --- tests for pretty printing keys ---

func TestTruncateWithEllipsis(t *testing.T) {
	digits := "1234567890"
	// do not truncate
	require.Equal(t, truncateWithEllipsis(digits, 10, true), digits)
	require.Equal(t, truncateWithEllipsis(digits, 10, false), digits)
	require.Equal(t, truncateWithEllipsis(digits, 11, true), digits)
	require.Equal(t, truncateWithEllipsis(digits, 11, false), digits)

	// left and right truncate
	require.Equal(t, truncateWithEllipsis(digits, 8, true), "...67890")
	require.Equal(t, truncateWithEllipsis(digits, 8, false), "12345...")
}

func TestKeyInfoSorter(t *testing.T) {
	expected := []keyInfo{
		{role: data.CanonicalRootRole, gun: "", keyID: "a", location: "i"},
		{role: data.CanonicalRootRole, gun: "", keyID: "a", location: "j"},
		{role: data.CanonicalRootRole, gun: "", keyID: "z", location: "z"},
		{role: "a", gun: "a", keyID: "a", location: "y"},
		{role: "b", gun: "a", keyID: "a", location: "y"},
		{role: "b", gun: "a", keyID: "b", location: "y"},
		{role: "b", gun: "a", keyID: "b", location: "z"},
		{role: "a", gun: "b", keyID: "a", location: "z"},
	}
	jumbled := make([]keyInfo, len(expected))
	// randomish indices
	for j, e := range []int{3, 6, 1, 4, 0, 7, 5, 2} {
		jumbled[j] = expected[e]
	}

	sort.Sort(keyInfoSorter(jumbled))
	require.True(t, reflect.DeepEqual(expected, jumbled),
		fmt.Sprintf("Expected %v, Got %v", expected, jumbled))
}

type otherMemoryStore struct {
	trustmanager.GenericKeyStore
}

func (l *otherMemoryStore) Name() string {
	return strings.Repeat("z", 70)
}

// If there are no keys in any of the key stores, a message that there are no
// signing keys should be displayed.
func TestPrettyPrintZeroKeys(t *testing.T) {
	ret := passphrase.ConstantRetriever("pass")
	emptyKeyStore := trustmanager.NewKeyMemoryStore(ret)

	var b bytes.Buffer
	prettyPrintKeys([]trustmanager.KeyStore{emptyKeyStore}, &b)
	text, err := ioutil.ReadAll(&b)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(text)), "\n")
	require.Len(t, lines, 1)
	require.Equal(t, "No signing keys found.", lines[0])
}

// Given a list of key stores, the keys should be pretty-printed with their
// roles, locations, IDs, and guns first in sorted order in the key store
func TestPrettyPrintRootAndSigningKeys(t *testing.T) {
	ret := passphrase.ConstantRetriever("pass")
	keyStores := []trustmanager.KeyStore{
		trustmanager.NewKeyMemoryStore(ret),
		&otherMemoryStore{GenericKeyStore: *trustmanager.NewKeyMemoryStore(ret)},
	}

	longNameShortened := "..." + strings.Repeat("z", 37)

	keys := make([]data.PrivateKey, 4)
	for i := 0; i < 4; i++ {
		key, err := utils.GenerateED25519Key(rand.Reader)
		require.NoError(t, err)
		keys[i] = key
	}

	root := data.CanonicalRootRole

	// add keys to the key stores
	require.NoError(t, keyStores[0].AddKey(trustmanager.KeyInfo{Role: root, Gun: ""}, keys[0]))
	require.NoError(t, keyStores[1].AddKey(trustmanager.KeyInfo{Role: root, Gun: ""}, keys[0]))
	require.NoError(t, keyStores[0].AddKey(trustmanager.KeyInfo{Role: data.CanonicalTargetsRole, Gun: data.GUN(strings.Repeat("/a", 30))}, keys[1]))
	require.NoError(t, keyStores[1].AddKey(trustmanager.KeyInfo{Role: data.CanonicalSnapshotRole, Gun: "short/gun"}, keys[1]))
	require.NoError(t, keyStores[0].AddKey(trustmanager.KeyInfo{Role: "targets/a", Gun: ""}, keys[3]))
	require.NoError(t, keyStores[0].AddKey(trustmanager.KeyInfo{Role: "invalidRole", Gun: ""}, keys[2]))

	expected := [][]string{
		// root always comes first
		{root.String(), keys[0].ID(), keyStores[0].Name()},
		{root.String(), keys[0].ID(), longNameShortened},
		// these have no gun, so they come first
		{"invalidRole", keys[2].ID(), keyStores[0].Name()},
		{"targets/a", keys[3].ID(), keyStores[0].Name()},
		// these have guns, and are sorted then by guns
		{data.CanonicalTargetsRole.String(), "..." + strings.Repeat("/a", 11), keys[1].ID(), keyStores[0].Name()},
		{data.CanonicalSnapshotRole.String(), "short/gun", keys[1].ID(), longNameShortened},
	}

	var b bytes.Buffer
	prettyPrintKeys(keyStores, &b)
	text, err := ioutil.ReadAll(&b)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(text)), "\n")
	require.Len(t, lines, len(expected)+2)

	// starts with headers
	require.True(t, reflect.DeepEqual(strings.Fields(lines[0]),
		[]string{"ROLE", "GUN", "KEY", "ID", "LOCATION"}))
	require.Equal(t, "----", lines[1][:4])

	for i, line := range lines[2:] {
		// we are purposely not putting spaces in test data so easier to split
		splitted := strings.Fields(line)
		for j, v := range splitted {
			require.Equal(t, expected[i][j], strings.TrimSpace(v))
		}
	}
}

// --- tests for pretty printing targets ---

// If there are no targets, no table is printed, only a line saying that there
// are no targets.
func TestPrettyPrintZeroTargets(t *testing.T) {
	var b bytes.Buffer
	prettyPrintTargets([]*client.TargetWithRole{}, &b)
	text, err := ioutil.ReadAll(&b)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(text)), "\n")
	require.Len(t, lines, 1)
	require.Equal(t, "No targets present in this repository.", lines[0])

}

// Targets are sorted by name, and the name, SHA256 digest, size, and role are
// printed.
func TestPrettyPrintSortedTargets(t *testing.T) {
	hashes := make([][]byte, 3)
	var err error
	for i, letter := range []string{"a012", "b012", "c012"} {
		hashes[i], err = hex.DecodeString(letter)
		require.NoError(t, err)
	}
	unsorted := []*client.TargetWithRole{
		{Target: client.Target{Name: "zebra", Hashes: data.Hashes{"sha256": hashes[0]}, Length: 8}, Role: "targets/b"},
		{Target: client.Target{Name: "aardvark", Hashes: data.Hashes{"sha256": hashes[1]}, Length: 1},
			Role: "targets"},
		{Target: client.Target{Name: "bee", Hashes: data.Hashes{"sha256": hashes[2]}, Length: 5}, Role: "targets/a"},
	}

	var b bytes.Buffer
	prettyPrintTargets(unsorted, &b)
	text, err := ioutil.ReadAll(&b)
	require.NoError(t, err)

	expected := [][]string{
		{"aardvark", "b012", "1", "targets"},
		{"bee", "c012", "5", "targets/a"},
		{"zebra", "a012", "8", "targets/b"},
	}

	lines := strings.Split(strings.TrimSpace(string(text)), "\n")
	require.Len(t, lines, len(expected)+2)

	// starts with headers
	require.True(t, reflect.DeepEqual(strings.Fields(lines[0]), strings.Fields(
		"NAME     DIGEST      SIZE (BYTES)   ROLE")))
	require.Equal(t, "----", lines[1][:4])

	for i, line := range lines[2:] {
		splitted := strings.Fields(line)
		require.Equal(t, expected[i], splitted)
	}
}

// --- tests for pretty printing roles ---

// If there are no roles, no table is printed, only a line saying that there
// are no roles.
func TestPrettyPrintZeroRoles(t *testing.T) {
	var b bytes.Buffer
	prettyPrintRoles([]data.Role{}, &b, "delegations")
	text, err := ioutil.ReadAll(&b)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(text)), "\n")
	require.Len(t, lines, 1)
	require.Equal(t, "No delegations present in this repository.", lines[0])
}

// Roles are sorted by name, and the name, paths, and KeyIDs are printed.
func TestPrettyPrintSortedRoles(t *testing.T) {
	var err error

	unsorted := []data.Role{
		{Name: "targets/zebra", Paths: []string{"stripes", "black", "white"}, RootRole: data.RootRole{KeyIDs: []string{"101"}, Threshold: 1}},
		{Name: "targets/aardvark/unicorn/pony", Paths: []string{"rainbows"}, RootRole: data.RootRole{KeyIDs: []string{"135"}, Threshold: 1}},
		{Name: "targets/bee", Paths: []string{"honey"}, RootRole: data.RootRole{KeyIDs: []string{"246"}, Threshold: 1}},
		{Name: "targets/bee/wasp", Paths: []string{"honey/sting", "stuff"}, RootRole: data.RootRole{KeyIDs: []string{"246", "468"}, Threshold: 1}},
	}

	var b bytes.Buffer
	prettyPrintRoles(unsorted, &b, "delegations")
	text, err := ioutil.ReadAll(&b)
	require.NoError(t, err)

	expected := [][]string{
		{"targets/aardvark/unicorn/pony", "rainbows", "135", "1"},
		{"targets/bee", "honey", "246", "1"},
		{"targets/bee/wasp", "honey/sting", "246", "1"},
		{"stuff", "468"}, // Extra keys and paths are printed to extra rows
		{"targets/zebra", "black", "101", "1"},
		{"stripes"},
		{"white"},
	}

	lines := strings.Split(strings.TrimSpace(string(text)), "\n")
	require.Len(t, lines, len(expected)+2)

	// starts with headers
	require.True(t, reflect.DeepEqual(strings.Fields(lines[0]), strings.Fields(
		"ROLE     PATHS      KEY IDS   THRESHOLD")))
	require.Equal(t, "----", lines[1][:4])

	for i, line := range lines[2:] {
		splitted := strings.Fields(line)
		require.Equal(t, expected[i], splitted)
	}
}
