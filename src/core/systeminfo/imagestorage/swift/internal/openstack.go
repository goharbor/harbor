package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

// IdentityHandler mocks KeyStone's catalog endpoint for tests purpose
func IdentityHandler(identityServer *httptest.Server, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
	{
		"versions": {
			"values": [
				{
					"status": "stable",
					"id": "v3.0",
					"links": [
						{ "href": "%s", "rel": "self" }
					]
				},
				{
					"status": "stable",
					"id": "v2.0",
					"links": [
						{ "href": "%s", "rel": "self" }
					]
				}
			]
		}
	}
`, identityServer.URL+"/v3/", identityServer.URL+"/v2.0/")
}

// CatalogHandler mocks KeyStone's token creation endpoint for tests purpose
func CatalogHandler(url string, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Subject-Token", "token")

	w.WriteHeader(http.StatusCreated)
	now := time.Now()
	fmt.Fprintf(
		w,
		`{
			"token": {
				"audit_ids": [],
				"catalog": [
					{
						"endpoints": [
							{
								"id": "1",
								"interface": "public",
								"region": "region",
								"url": "%s"
							}
						],
						"id": "1",
						"type": "object-store",
						"name": "Swift"
					}
				],
				"expires_at": "%s",
				"is_domain": false,
				"issued_at": "%s",
				"methods": [
					"password"
				],
				"project": {
					"domain": {
						"id": "1",
						"name": "domain"
					},
					"id": "1",
					"name": "project"
				},
				"roles": [],
				"service_providers": [],
				"user": {}
			}
		}`,
		url,
		now.Add(10*time.Minute),
		now,
	)
}

// EmptyCatalogHandler mocks KeyStone's token creation endpoint with an empty catalog for tests purpose
func EmptyCatalogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Subject-Token", "token")

	w.WriteHeader(http.StatusCreated)
	now := time.Now()
	fmt.Fprintf(
		w,
		`{
			"token": {
				"audit_ids": [],
				"catalog": [],
				"expires_at": "%s",
				"is_domain": false,
				"issued_at": "%s",
				"methods": [
					"password"
				],
				"project": {
					"domain": {
						"id": "1",
						"name": "domain"
					},
					"id": "1",
					"name": "project"
				},
				"roles": [],
				"service_providers": [],
				"user": {}
			}
		}`,
		now.Add(10*time.Minute),
		now,
	)
}
