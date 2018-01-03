package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/notary"
	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/require"
)

// the default location for the config file is in ~/.notary/config.json - even if it doesn't exist.
func TestNotaryConfigFileDefault(t *testing.T) {
	commander := &notaryCommander{
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
	}

	config, err := commander.parseConfig()
	require.NoError(t, err)
	configFileUsed := config.ConfigFileUsed()
	require.True(t, strings.HasSuffix(configFileUsed,
		filepath.Join(".notary", "config.json")), "Unknown config file: %s", configFileUsed)
}

// the default server address is notary-server
func TestRemoteServerDefault(t *testing.T) {
	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)
	configFile := filepath.Join(tempDir, "config.json")

	commander := &notaryCommander{
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
	}

	// set a blank config file, so it doesn't check ~/.notary/config.json by default
	// and execute a random command so that the flags are parsed
	cmd := commander.GetCommand()
	cmd.SetArgs([]string{"-c", configFile, "list"})
	cmd.SetOutput(new(bytes.Buffer)) // eat the output
	cmd.Execute()

	config, err := commander.parseConfig()
	require.NoError(t, err)
	require.Equal(t, "https://notary-server:4443", getRemoteTrustServer(config))
}

// providing a config file uses the config file's server url instead
func TestRemoteServerUsesConfigFile(t *testing.T) {
	tempDir := tempDirWithConfig(t, `{"remote_server": {"url": "https://myserver"}}`)
	defer os.RemoveAll(tempDir)
	configFile := filepath.Join(tempDir, "config.json")

	commander := &notaryCommander{
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
	}

	// set a config file, so it doesn't check ~/.notary/config.json by default,
	// and execute a random command so that the flags are parsed
	cmd := commander.GetCommand()
	cmd.SetArgs([]string{"-c", configFile, "list"})
	cmd.SetOutput(new(bytes.Buffer)) // eat the output
	cmd.Execute()

	config, err := commander.parseConfig()
	require.NoError(t, err)
	require.Equal(t, "https://myserver", getRemoteTrustServer(config))
}

// a command line flag overrides the config file's server url
func TestRemoteServerCommandLineFlagOverridesConfig(t *testing.T) {
	tempDir := tempDirWithConfig(t, `{"remote_server": {"url": "https://myserver"}}`)
	defer os.RemoveAll(tempDir)
	configFile := filepath.Join(tempDir, "config.json")

	commander := &notaryCommander{
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
	}

	// set a config file, so it doesn't check ~/.notary/config.json by default,
	// and execute a random command so that the flags are parsed
	cmd := commander.GetCommand()
	cmd.SetArgs([]string{"-c", configFile, "-s", "http://overridden", "list"})
	cmd.SetOutput(new(bytes.Buffer)) // eat the output
	cmd.Execute()

	config, err := commander.parseConfig()
	require.NoError(t, err)
	require.Equal(t, "http://overridden", getRemoteTrustServer(config))
}

// invalid commands for `notary addhash`
func TestInvalidAddHashCommands(t *testing.T) {
	tempDir := tempDirWithConfig(t, `{"remote_server": {"url": "https://myserver"}}`)
	defer os.RemoveAll(tempDir)
	configFile := filepath.Join(tempDir, "config.json")

	b := new(bytes.Buffer)
	cmd := NewNotaryCommand()
	cmd.SetOutput(b)

	// No hashes given
	cmd.SetArgs(append([]string{"-c", configFile, "-d", tempDir}, "addhash", "gun", "test", "10"))
	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Must specify a GUN, target, byte size of target data, and at least one hash")

	// Invalid byte size given
	cmd = NewNotaryCommand()
	cmd.SetArgs(append([]string{"-c", configFile, "-d", tempDir}, "addhash", "gun", "test", "sizeNotAnInt", "--sha256", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	err = cmd.Execute()
	require.Error(t, err)

	// Invalid sha256 size given
	cmd = NewNotaryCommand()
	cmd.SetArgs(append([]string{"-c", configFile, "-d", tempDir}, "addhash", "gun", "test", "1", "--sha256", "a"))
	err = cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid sha256 hex contents provided")

	// Invalid sha256 hex given
	cmd = NewNotaryCommand()
	cmd.SetArgs(append([]string{"-c", configFile, "-d", tempDir}, "addhash", "gun", "test", "1", "--sha256", "***aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa***"))
	err = cmd.Execute()
	require.Error(t, err)

	// Invalid sha512 size given
	cmd = NewNotaryCommand()
	cmd.SetArgs(append([]string{"-c", configFile, "-d", tempDir}, "addhash", "gun", "test", "1", "--sha512", "a"))
	err = cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid sha512 hex contents provided")

	// Invalid sha512 hex given
	cmd = NewNotaryCommand()
	cmd.SetArgs(append([]string{"-c", configFile, "-d", tempDir}, "addhash", "gun", "test", "1", "--sha512", "***aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa******aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa***"))
	err = cmd.Execute()
	require.Error(t, err)
}

var exampleValidCommands = []string{
	"init repo",
	"list repo",
	"status repo",
	"reset repo --all",
	"publish repo",
	"add repo v1 somefile",
	"addhash repo targetv1 --sha256 aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa 10",
	"verify repo v1",
	"key list",
	"key rotate repo snapshot",
	"key generate rsa",
	"key remove e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	"key passwd e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	"key import backup.pem",
	"delegation list repo",
	"delegation add repo targets/releases path/to/pem/file.pem",
	"delegation remove repo targets/releases",
	"witness gun targets/releases",
	"delete repo",
}

// config parsing bugs are propagated in all commands
func TestConfigParsingErrorsPropagatedByCommands(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "empty-dir")
	require.NoError(t, err)
	defer os.RemoveAll(tempdir)

	for _, args := range exampleValidCommands {
		b := new(bytes.Buffer)
		cmd := NewNotaryCommand()
		cmd.SetOutput(b)

		cmd.SetArgs(append(
			[]string{"-c", filepath.Join(tempdir, "idonotexist.json"), "-d", tempdir},
			strings.Fields(args)...))
		err = cmd.Execute()

		require.Error(t, err, "expected error when running `notary %s`", args)
		require.Contains(t, err.Error(), "error opening config file", "running `notary %s`", args)
		require.NotContains(t, b.String(), "Usage:")
	}
}

// insufficient arguments produce an error before any parsing of configs happens
func TestInsufficientArgumentsReturnsErrorAndPrintsUsage(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "empty-dir")
	require.NoError(t, err)
	defer os.RemoveAll(tempdir)

	for _, args := range exampleValidCommands {
		b := new(bytes.Buffer)
		cmd := NewNotaryCommand()
		cmd.SetOutput(b)

		arglist := strings.Fields(args)
		if args == "key list" || args == "key generate rsa" {
			// in these case, "key" or "key generate" are valid commands, so add an arg to them instead
			arglist = append(arglist, "extraArg")
		} else {
			arglist = arglist[:len(arglist)-1]
		}

		invalid := strings.Join(arglist, " ")

		cmd.SetArgs(append(
			[]string{"-c", filepath.Join(tempdir, "idonotexist.json"), "-d", tempdir}, arglist...))
		err = cmd.Execute()

		require.NotContains(t, err.Error(), "error opening config file", "running `notary %s`", invalid)
		// it's a usage error, so the usage is printed
		require.Contains(t, b.String(), "Usage:", "expected usage when running `notary %s`", invalid)
	}
}

// The bare notary command and bare subcommands all print out usage
func TestBareCommandPrintsUsageAndNoError(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "empty-dir")
	require.NoError(t, err)
	defer os.RemoveAll(tempdir)

	// just the notary command
	b := new(bytes.Buffer)
	cmd := NewNotaryCommand()
	cmd.SetOutput(b)

	cmd.SetArgs([]string{"-c", filepath.Join(tempdir, "idonotexist.json")})
	require.NoError(t, cmd.Execute(), "Expected no error from a help request")
	// usage is printed
	require.Contains(t, b.String(), "Usage:", "expected usage when running `notary`")

	// notary key and notary delegation
	for _, bareCommand := range []string{"key", "delegation"} {
		b := new(bytes.Buffer)
		cmd := NewNotaryCommand()
		cmd.SetOutput(b)

		cmd.SetArgs([]string{"-c", filepath.Join(tempdir, "idonotexist.json"), "-d", tempdir, bareCommand})
		require.NoError(t, cmd.Execute(), "Expected no error from a help request")
		// usage is printed
		require.Contains(t, b.String(), "Usage:", "expected usage when running `notary %s`", bareCommand)
	}
}

type recordingMetaStore struct {
	gotten []string
	storage.MemStorage
}

// GetCurrent gets the metadata from the underlying MetaStore, but also records
// that the metadata was requested
func (r *recordingMetaStore) GetCurrent(gun data.GUN, role data.RoleName) (*time.Time, []byte, error) {
	r.gotten = append(r.gotten, fmt.Sprintf("%s.%s", gun.String(), role.String()))
	return r.MemStorage.GetCurrent(gun, role)
}

// GetChecksum gets the metadata from the underlying MetaStore, but also records
// that the metadata was requested
func (r *recordingMetaStore) GetChecksum(gun data.GUN, role data.RoleName, checksum string) (*time.Time, []byte, error) {
	r.gotten = append(r.gotten, fmt.Sprintf("%s.%s", gun.String(), role.String()))
	return r.MemStorage.GetChecksum(gun, role, checksum)
}

// the config can provide all the TLS information necessary - the root ca file,
// the tls client files - they are all relative to the directory of the config
// file, and not the cwd
func TestConfigFileTLSCannotBeRelativeToCWD(t *testing.T) {
	// Set up server that with a self signed cert
	var err error
	// add a handler for getting the root
	m := &recordingMetaStore{MemStorage: *storage.NewMemStorage()}
	s := httptest.NewUnstartedServer(setupServerHandler(m))
	s.TLS, err = tlsconfig.Server(tlsconfig.Options{
		CertFile:   "../../fixtures/notary-server.crt",
		KeyFile:    "../../fixtures/notary-server.key",
		CAFile:     "../../fixtures/root-ca.crt",
		ClientAuth: tls.RequireAndVerifyClientCert,
	})
	require.NoError(t, err)
	s.StartTLS()
	defer s.Close()

	// test that a config file with certs that are relative to the cwd fail
	tempDir := tempDirWithConfig(t, fmt.Sprintf(`{
		"remote_server": {
			"url": "%s",
			"root_ca": "../../fixtures/root-ca.crt",
			"tls_client_cert": "../../fixtures/notary-server.crt",
			"tls_client_key": "../../fixtures/notary-server.key"
		}
	}`, s.URL))
	defer os.RemoveAll(tempDir)
	configFile := filepath.Join(tempDir, "config.json")

	// set a config file, so it doesn't check ~/.notary/config.json by default,
	// and execute a random command so that the flags are parsed
	cmd := NewNotaryCommand()
	cmd.SetArgs([]string{"-c", configFile, "-d", tempDir, "list", "repo"})
	cmd.SetOutput(new(bytes.Buffer)) // eat the output
	err = cmd.Execute()
	require.Error(t, err, "expected a failure due to TLS")
	require.Contains(t, err.Error(), "TLS", "should have been a TLS error")

	// validate that we failed to connect and attempt any downloads at all
	require.Len(t, m.gotten, 0)
}

// the config can provide all the TLS information necessary - the root ca file,
// the tls client files - they are all relative to the directory of the config
// file, and not the cwd, or absolute paths
func TestConfigFileTLSCanBeRelativeToConfigOrAbsolute(t *testing.T) {
	// Set up server that with a self signed cert
	var err error
	// add a handler for getting the root
	m := &recordingMetaStore{MemStorage: *storage.NewMemStorage()}
	s := httptest.NewUnstartedServer(setupServerHandler(m))
	s.TLS, err = tlsconfig.Server(tlsconfig.Options{
		CertFile:   "../../fixtures/notary-server.crt",
		KeyFile:    "../../fixtures/notary-server.key",
		CAFile:     "../../fixtures/root-ca.crt",
		ClientAuth: tls.RequireAndVerifyClientCert,
	})
	require.NoError(t, err)
	s.StartTLS()
	defer s.Close()

	tempDir, err := ioutil.TempDir("", "config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	configFile, err := os.Create(filepath.Join(tempDir, "config.json"))
	require.NoError(t, err)
	fmt.Fprintf(configFile, `{
		"remote_server": {
			"url": "%s",
			"root_ca": "root-ca.crt",
			"tls_client_cert": %s,
			"tls_client_key": "notary-server.key"
		}
	}`, s.URL, strconv.Quote(filepath.Join(tempDir, "notary-server.crt")))
	configFile.Close()

	// copy the certs to be relative to the config directory
	for _, fname := range []string{"notary-server.crt", "notary-server.key", "root-ca.crt"} {
		content, err := ioutil.ReadFile(filepath.Join("../../fixtures", fname))
		require.NoError(t, err)
		require.NoError(t, ioutil.WriteFile(filepath.Join(tempDir, fname), content, 0766))
	}

	// set a config file, so it doesn't check ~/.notary/config.json by default,
	// and execute a random command so that the flags are parsed
	cmd := NewNotaryCommand()
	cmd.SetArgs([]string{"-c", configFile.Name(), "-d", tempDir, "list", "repo"})
	cmd.SetOutput(new(bytes.Buffer)) // eat the output
	err = cmd.Execute()
	require.Error(t, err, "there was no repository, so list should have failed")
	require.NotContains(t, err.Error(), "TLS", "there was no TLS error though!")

	// validate that we actually managed to connect and attempted to download the root though
	require.Len(t, m.gotten, 1)
	require.Equal(t, m.gotten[0], "repo.root")
}

// Whatever TLS config is in the config file can be overridden by the command line
// TLS flags, which are relative to the CWD (not the config) or absolute
func TestConfigFileOverridenByCmdLineFlags(t *testing.T) {
	// Set up server that with a self signed cert
	var err error
	// add a handler for getting the root
	m := &recordingMetaStore{MemStorage: *storage.NewMemStorage()}
	s := httptest.NewUnstartedServer(setupServerHandler(m))
	s.TLS, err = tlsconfig.Server(tlsconfig.Options{
		CertFile:   "../../fixtures/notary-server.crt",
		KeyFile:    "../../fixtures/notary-server.key",
		CAFile:     "../../fixtures/root-ca.crt",
		ClientAuth: tls.RequireAndVerifyClientCert,
	})
	require.NoError(t, err)
	s.StartTLS()
	defer s.Close()

	tempDir := tempDirWithConfig(t, fmt.Sprintf(`{
		"remote_server": {
			"url": "%s",
			"root_ca": "nope",
			"tls_client_cert": "nope",
			"tls_client_key": "nope"
		}
	}`, s.URL))
	defer os.RemoveAll(tempDir)
	configFile := filepath.Join(tempDir, "config.json")

	// set a config file, so it doesn't check ~/.notary/config.json by default,
	// and execute a random command so that the flags are parsed
	cwd, err := os.Getwd()
	require.NoError(t, err)

	cmd := NewNotaryCommand()
	cmd.SetArgs([]string{
		"-c", configFile, "-d", tempDir, "list", "repo",
		"--tlscacert", "../../fixtures/root-ca.crt",
		"--tlscert", filepath.Clean(filepath.Join(cwd, "../../fixtures/notary-server.crt")),
		"--tlskey", "../../fixtures/notary-server.key"})
	cmd.SetOutput(new(bytes.Buffer)) // eat the output
	err = cmd.Execute()
	require.Error(t, err, "there was no repository, so list should have failed")
	require.NotContains(t, err.Error(), "TLS", "there was no TLS error though!")

	// validate that we actually managed to connect and attempted to download the root though
	require.Len(t, m.gotten, 1)
	require.Equal(t, m.gotten[0], "repo.root")
}

// the config can specify trust pinning settings for TOFUs, as well as pinned Certs or CA
func TestConfigFileTrustPinning(t *testing.T) {
	var err error

	tempDir := tempDirWithConfig(t, `{
        "trust_pinning": {
            "disable_tofu": false
         }
	}`)
	defer os.RemoveAll(tempDir)
	commander := &notaryCommander{
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
		configFile:   filepath.Join(tempDir, "config.json"),
	}

	// Check that tofu was set correctly
	config, err := commander.parseConfig()
	require.NoError(t, err)
	require.Equal(t, false, config.GetBool("trust_pinning.disable_tofu"))
	trustPin, err := getTrustPinning(config)
	require.NoError(t, err)
	require.Equal(t, false, trustPin.DisableTOFU)

	tempDir = tempDirWithConfig(t, `{
		"remote_server": {
			"url": "%s"
		},
		"trust_pinning": {
		    "disable_tofu": true
		}
	}`)
	defer os.RemoveAll(tempDir)
	commander = &notaryCommander{
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
		configFile:   filepath.Join(tempDir, "config.json"),
	}

	// Check that tofu was correctly disabled
	config, err = commander.parseConfig()
	require.NoError(t, err)
	require.Equal(t, true, config.GetBool("trust_pinning.disable_tofu"))
	trustPin, err = getTrustPinning(config)
	require.NoError(t, err)
	require.Equal(t, true, trustPin.DisableTOFU)

	tempDir = tempDirWithConfig(t, fmt.Sprintf(`{
		"trust_pinning": {
		    "certs": {
		        "repo3": ["%s"]
		    }
		 }
	}`, strings.Repeat("x", notary.SHA256HexSize)))
	defer os.RemoveAll(tempDir)
	commander = &notaryCommander{
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
		configFile:   filepath.Join(tempDir, "config.json"),
	}

	config, err = commander.parseConfig()
	require.NoError(t, err)
	require.Equal(t, []interface{}{strings.Repeat("x", notary.SHA256HexSize)}, config.GetStringMap("trust_pinning.certs")["repo3"])
	trustPin, err = getTrustPinning(config)
	require.NoError(t, err)
	require.Equal(t, strings.Repeat("x", notary.SHA256HexSize), trustPin.Certs["repo3"][0])

	// Check that an invalid cert ID pinning format fails
	tempDir = tempDirWithConfig(t, fmt.Sprintf(`{
		"trust_pinning": {
		    "certs": {
		        "repo3": "%s"
		    }
		 }
	}`, strings.Repeat("x", notary.SHA256HexSize)))
	defer os.RemoveAll(tempDir)
	commander = &notaryCommander{
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
		configFile:   filepath.Join(tempDir, "config.json"),
	}

	config, err = commander.parseConfig()
	require.NoError(t, err)
	trustPin, err = getTrustPinning(config)
	require.Error(t, err)

	tempDir = tempDirWithConfig(t, fmt.Sprintf(`{
		"trust_pinning": {
		    "ca": {
		        "repo4": "%s"
		    }
		 }
	}`, "root-ca.crt"))
	defer os.RemoveAll(tempDir)
	commander = &notaryCommander{
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
		configFile:   filepath.Join(tempDir, "config.json"),
	}

	config, err = commander.parseConfig()
	require.NoError(t, err)
	require.Equal(t, "root-ca.crt", config.GetStringMap("trust_pinning.ca")["repo4"])
	trustPin, err = getTrustPinning(config)
	require.NoError(t, err)
	require.Equal(t, "root-ca.crt", trustPin.CA["repo4"])
}

// sets the env vars to empty, and returns a function to reset them at the end
func cleanupAndSetEnvVars() func() {
	orig := map[string]string{
		"NOTARY_ROOT_PASSPHRASE":       "",
		"NOTARY_TARGETS_PASSPHRASE":    "",
		"NOTARY_SNAPSHOT_PASSPHRASE":   "",
		"NOTARY_DELEGATION_PASSPHRASE": "",
	}
	for envVar := range orig {
		orig[envVar] = os.Getenv(envVar)
		os.Setenv(envVar, "")
	}

	return func() {
		for envVar, value := range orig {
			if value == "" {
				os.Unsetenv(envVar)
			} else {
				os.Setenv(envVar, value)
			}
		}
	}
}

func TestPassphraseRetrieverCaching(t *testing.T) {
	defer cleanupAndSetEnvVars()()
	// Only set up one passphrase environment var first for root
	require.NoError(t, os.Setenv("NOTARY_ROOT_PASSPHRASE", "root_passphrase"))

	// Check that root is cached
	retriever := getPassphraseRetriever()
	passphrase, giveup, err := retriever("key", data.CanonicalRootRole.String(), false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "root_passphrase")

	_, _, err = retriever("key", "user", false, 0)
	require.Error(t, err)
	_, _, err = retriever("key", data.CanonicalTargetsRole.String(), false, 0)
	require.Error(t, err)
	_, _, err = retriever("key", data.CanonicalSnapshotRole.String(), false, 0)
	require.Error(t, err)
	_, _, err = retriever("key", "targets/delegation", false, 0)
	require.Error(t, err)

	// Set up the rest of them
	require.NoError(t, os.Setenv("NOTARY_TARGETS_PASSPHRASE", "targets_passphrase"))
	require.NoError(t, os.Setenv("NOTARY_SNAPSHOT_PASSPHRASE", "snapshot_passphrase"))
	require.NoError(t, os.Setenv("NOTARY_DELEGATION_PASSPHRASE", "delegation_passphrase"))

	// Get a new retriever and check the caching
	retriever = getPassphraseRetriever()
	passphrase, giveup, err = retriever("key", data.CanonicalRootRole.String(), false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "root_passphrase")

	passphrase, giveup, err = retriever("key", data.CanonicalTargetsRole.String(), false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "targets_passphrase")

	passphrase, giveup, err = retriever("key", data.CanonicalSnapshotRole.String(), false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "snapshot_passphrase")

	passphrase, giveup, err = retriever("key", "targets/releases", false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "delegation_passphrase")

	// We don't require a targets/ prefix in PEM headers for delegation keys
	passphrase, giveup, err = retriever("key", "user", false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "delegation_passphrase")
}

func TestPassphraseRetrieverDelegationRoleCaching(t *testing.T) {
	defer cleanupAndSetEnvVars()()
	// Only set up one passphrase environment var first for delegations
	require.NoError(t, os.Setenv("NOTARY_DELEGATION_PASSPHRASE", "delegation_passphrase"))

	// Check that any delegation role is cached
	retriever := getPassphraseRetriever()

	passphrase, giveup, err := retriever("key", "targets/releases", false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "delegation_passphrase")
	passphrase, giveup, err = retriever("key", "targets/delegation", false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "delegation_passphrase")
	passphrase, giveup, err = retriever("key", "targets/a/b/c/d", false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "delegation_passphrase")

	// Also check arbitrary usernames that are non-BaseRoles or imported so that this can be shared across keys
	passphrase, giveup, err = retriever("key", "user", false, 0)
	require.NoError(t, err)
	require.False(t, giveup)
	require.Equal(t, passphrase, "delegation_passphrase")

	// Make sure base roles fail
	_, _, err = retriever("key", data.CanonicalRootRole.String(), false, 0)
	require.Error(t, err)
	_, _, err = retriever("key", data.CanonicalTargetsRole.String(), false, 0)
	require.Error(t, err)
	_, _, err = retriever("key", data.CanonicalSnapshotRole.String(), false, 0)
	require.Error(t, err)
}
