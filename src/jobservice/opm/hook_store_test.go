// Copyright 2018 The Harbor Authors. All rights reserved.
package opm

import "testing"

func TestHookStore(t *testing.T) {
	store := NewHookStore()

	reportURL := "http://localhost:9090/report"
	store.Add("id_1", reportURL)
	url, ok := store.Get("id_1")
	if !ok || url != reportURL {
		t.Errorf("expect hook url '%s' but got '%s'", reportURL, url)
	}
	u, ok := store.Remove("id_1")
	if !ok || u != reportURL {
		t.Errorf("expect deleted '%s' but failed to do", reportURL)
	}
}
