package lib

import (
	"testing"
)

var testcases = []struct {
	url         string
	expectedUrl string
	valid       bool
}{
	{"http://harbor.foo.com", "http://harbor.foo.com", true},
	{"http://harbor.foo.com/", "http://harbor.foo.com", true},
	{"http://harbor.foo.com/path", "http://harbor.foo.com/path", true},
	{"/", "", false},
	{"foo.html", "http://foo.html", true},
	{"*", "http://*", true},
	{"http://127.0.0.1/", "http://127.0.0.1", true},
	{"http://127.0.0.1:8080/", "http://127.0.0.1:8080", true},
	{"http://[fe80::1]/", "http://[fe80::1]", true},
	{"http://[fe80::1]:8080/", "http://[fe80::1]:8080", true},

	{"http://[fe80::1%25en0]/", "http://[fe80::1%en0]", true},
	{"http://[fe80::1%25en0]:8080/", "http://[fe80::1%en0]:8080", true},
	{"http://[fe80::1%25%65%6e%301-._~]/", "http://[fe80::1%en01-._~]", true},
	{"http://[fe80::1%25%65%6e%301-._~]:8080/", "http://[fe80::1%en01-._~]:8080", true},

	{"http://127.0.0.%31/", "", false},
	{"http://127.0.0.%31:8080/", "", false},
	{"http://10.0.0.1/test.txt#/api/version", "http://10.0.0.1/test.txt", true},
}

func TestValidateHTTPURL(t *testing.T) {
	for _, test := range testcases {
		url, err := ValidateHTTPURL(test.url)
		if test.valid {
			if err != nil {
				t.Errorf("ValidateHTTPURL:%q gave err %v; want no error", test.url, err)
			}
			if url != test.expectedUrl {
				t.Errorf("ValidateHTTPURL:%q gave %s; want %s", test.url, url, test.expectedUrl)
			}
		} else if !test.valid && err == nil {
			t.Errorf("ValidateHTTPURL:%q gave <nil> error; want some error", test.url)
		}
	}

}
