package passphrase

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/notary"
	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/require"
)

func assertAskOnceForKey(t *testing.T, in, out *bytes.Buffer, retriever notary.PassRetriever, password, role string) {
	_, err := in.WriteString(password + "\n")
	require.NoError(t, err)

	pass, giveUp, err := retriever("repo/0123456789abcdef", role, false, 0)
	require.NoError(t, err)
	require.False(t, giveUp)
	require.Equal(t, password, pass)

	text, err := ioutil.ReadAll(out)
	require.NoError(t, err)
	require.Equal(t, "Enter passphrase for "+role+" key with ID 0123456 (repo):",
		strings.TrimSpace(string(text)))
}

// PromptRetrieverWithInOut prompts for delegations passwords (non creation) if needed
func TestGetPassphraseForUsingDelegationKey(t *testing.T) {
	var in bytes.Buffer
	var out bytes.Buffer

	retriever := PromptRetrieverWithInOut(&in, &out, nil)

	for i := 0; i < 3; i++ {
		target := fmt.Sprintf("targets/level%d", i)
		password := fmt.Sprintf("randompass%d", i)

		assertAskOnceForKey(t, &in, &out, retriever, password, target)
	}
}

// PromptRetrieverWithInOut prompts for passwords up to 10 times when creating
func TestGetPassphraseLimitsShortPassphrases(t *testing.T) {
	var in bytes.Buffer
	var out bytes.Buffer

	retriever := PromptRetrieverWithInOut(&in, &out, nil)

	repeatedShortPass := strings.Repeat("a\n", 22)
	_, err := in.WriteString(repeatedShortPass)
	require.NoError(t, err)

	_, _, err = retriever("randomRepo", "targets/randomRole", true, 0)
	require.Error(t, err)
	require.IsType(t, ErrTooManyAttempts, err)
}

// PromptRetrieverWithInOut prompts for passwords up to 10 times when creating
func TestGetPassphraseLimitsMismatchingPassphrases(t *testing.T) {
	var in bytes.Buffer
	var out bytes.Buffer

	retriever := PromptRetrieverWithInOut(&in, &out, nil)

	repeatedShortPass := strings.Repeat("password\nmismatchingpass\n", 11)
	_, err := in.WriteString(repeatedShortPass)
	require.NoError(t, err)

	_, _, err = retriever("randomRepo", "targets/randomRole", true, 0)
	require.Error(t, err)
	require.IsType(t, ErrTooManyAttempts, err)
}

// PromptRetrieverWithInOut prompts for creating delegations passwords if needed
func TestGetPassphraseForCreatingDelegationKey(t *testing.T) {
	var in bytes.Buffer
	var out bytes.Buffer

	retriever := PromptRetrieverWithInOut(&in, &out, nil)

	_, err := in.WriteString("passphrase\npassphrase\n")
	require.NoError(t, err)

	pass, giveUp, err := retriever("repo/0123456789abcdef", "targets/a", true, 0)
	require.NoError(t, err)
	require.False(t, giveUp)
	require.Equal(t, "passphrase", pass)

	text, err := ioutil.ReadAll(&out)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(text)), "\n")

	expectedText := []string{
		`Enter passphrase for new targets/a key with ID 0123456 (repo): `,
		`Repeat passphrase for new targets/a key with ID 0123456 (repo):`,
	}

	require.Equal(t, expectedText, lines)
}

// PromptRetrieverWithInOut, if asked for root, targets, snapshot, and delegation
// passphrases in that order will cache each of the keys except for the delegation key
func TestRolePromptingAndCaching(t *testing.T) {
	var in bytes.Buffer
	var out bytes.Buffer

	retriever := PromptRetrieverWithInOut(&in, &out, nil)

	assertAskOnceForKey(t, &in, &out, retriever, "rootpassword", data.CanonicalRootRole.String())
	assertAskOnceForKey(t, &in, &out, retriever, "targetspassword", data.CanonicalTargetsRole.String())
	assertAskOnceForKey(t, &in, &out, retriever, "snapshotpassword", data.CanonicalSnapshotRole.String())
	assertAskOnceForKey(t, &in, &out, retriever, "delegationpass", "targets/delegation")

	// ask for root password, but it should already be cached
	pass, giveUp, err := retriever("repo/0123456789abcdef", data.CanonicalRootRole.String(), false, 0)
	require.NoError(t, err)
	require.False(t, giveUp)
	require.Equal(t, "rootpassword", pass)

	// ask for targets password, but it should already be cached
	pass, giveUp, err = retriever("repo/0123456789abcdef", data.CanonicalTargetsRole.String(), false, 0)
	require.NoError(t, err)
	require.False(t, giveUp)
	require.Equal(t, "targetspassword", pass)

	// ask for snapshot password, but it should already be cached
	pass, giveUp, err = retriever("repo/0123456789abcdef", data.CanonicalSnapshotRole.String(), false, 0)
	require.NoError(t, err)
	require.False(t, giveUp)
	require.Equal(t, "snapshotpassword", pass)

	// ask for targets/delegation password, but it should already be cached
	pass, giveUp, err = retriever("repo/0123456789abcdef", "targets/delegation", false, 0)
	require.NoError(t, err)
	require.False(t, giveUp)
	require.Equal(t, "delegationpass", pass)

	// ask for different delegation password, which should not be cached
	_, _, err = retriever("repo/0123456789abcdef", "targets/delegation/new", false, 0)
	require.Error(t, err)
	text, err := ioutil.ReadAll(&out)
	require.NoError(t, err)
	require.Contains(t, string(text), "Enter passphrase for targets/delegation/new key with ID 0123456 (repo):")
}

// TestPromptRetrieverNeedsTerminal checks that PromptRetriever errors when not run with a terminal stdin
func TestPromptRetrieverNeedsTerminal(t *testing.T) {
	prompt := PromptRetriever()
	_, _, err := prompt("repo/0123456789abcdef", "targets/delegation/new", false, 0)
	require.Error(t, err)
	require.IsType(t, ErrNoInput, err)
}

// TestGetPassphrase checks getting passphrase from stdin
func TestGetPassphrase(t *testing.T) {
	var in bytes.Buffer

	_, err := in.WriteString("passphrase\n")
	require.NoError(t, err)

	stdin := bufio.NewReader(&in)
	passphrase, err := GetPassphrase(stdin)
	require.NoError(t, err)
	require.Equal(t, string(passphrase), "passphrase\n")
}
