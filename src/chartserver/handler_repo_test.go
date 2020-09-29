package chartserver

import "testing"

// Test get /index.yaml
func TestGetIndexFile(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	namespaces := []string{"repo1", "repo2"}
	indexFile, err := c.GetIndexFile(namespaces)
	if err != nil {
		t.Fatal(err)
	}

	if len(indexFile.Entries) != 5 {
		t.Fatalf("Expect index file with 5 entries, but got %d", len(indexFile.Entries))
	}

	_, ok := indexFile.Entries["repo1/harbor"]
	if !ok {
		t.Fatal("Expect chart entry 'repo1/harbor' but got nothing")
	}

	_, ok = indexFile.Entries["repo2/harbor"]
	if !ok {
		t.Fatal("Expect chart entry 'repo2/harbor' but got nothing")
	}
}
