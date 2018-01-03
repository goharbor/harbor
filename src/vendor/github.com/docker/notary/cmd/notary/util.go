package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

const (
	// The help text of auto publish
	htAutoPublish string = "Automatically attempt to publish after staging the change. Will also publish existing staged changes."
)

// getPayload is a helper function to get the content used to be verified
// either from an existing file or STDIN.
func getPayload(t *tufCommander) ([]byte, error) {

	// Reads from the given file
	if t.input != "" {
		// Please note that ReadFile will cut off the size if it was over 1e9.
		// Thus, if the size of the file exceeds 1GB, the over part will not be
		// loaded into the buffer.
		payload, err := ioutil.ReadFile(t.input)
		if err != nil {
			return nil, err
		}
		return payload, nil
	}

	// Reads all of the data on STDIN
	payload, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("Error reading content from STDIN: %v", err)
	}
	return payload, nil
}

// feedback is a helper function to print the payload to a file or STDOUT or keep quiet
// due to the value of flag "quiet" and "output".
func feedback(t *tufCommander, payload []byte) error {
	// We only get here when everything goes well, since the flag "quiet" was
	// provided, we output nothing but just return.
	if t.quiet {
		return nil
	}

	// Flag "quiet" was not "true", that's why we get here.
	if t.output != "" {
		return ioutil.WriteFile(t.output, payload, 0644)
	}

	os.Stdout.Write(payload)
	return nil
}
