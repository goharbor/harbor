package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPayload(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test-get-payload")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	file, err := os.Create(filepath.Join(tempDir, "content.txt"))
	require.NoError(t, err)

	fmt.Fprintf(file, "Release date: June 10, 2016 - Director: Duncan Jones")
	file.Close()

	commander := &tufCommander{
		input: file.Name(),
	}

	payload, err := getPayload(commander)
	require.NoError(t, err)
	require.Equal(t, "Release date: June 10, 2016 - Director: Duncan Jones", string(payload))
}

func TestFeedback(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test-feedback")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	file, err := os.Create(filepath.Join(tempDir, "content.txt"))
	require.NoError(t, err)

	// Expect it to print nothing since "quiet" takes priority.
	commander := &tufCommander{
		output: file.Name(),
		quiet:  true,
	}

	payload := []byte("Release date: June 10, 2016 - Director: Duncan Jones")
	err = feedback(commander, payload)
	require.NoError(t, err)

	content, err := ioutil.ReadFile(file.Name())
	require.NoError(t, err)
	require.Equal(t, "", string(content))
}
