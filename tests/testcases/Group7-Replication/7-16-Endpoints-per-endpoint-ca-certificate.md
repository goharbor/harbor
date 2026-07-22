Test 7-16 Endpoints per-endpoint CA certificate
=======

# Purpose

To verify admin user can configure a per-endpoint CA certificate for a registry
endpoint that uses a self-signed (or otherwise untrusted) certificate, and that
Harbor uses it to verify the connection instead of requiring "Verify Remote
Cert" to be disabled.

# References:

User guide

# Environments:

* This test requires at least two Harbor instances running and available,
  where the remote instance serves HTTPS with a self-signed certificate.

# Test Steps:

1. Login UI as admin user.
2. In `Administration->Registries` page, add an endpoint with a valid URL
   (HTTPS with self-signed certificate), username and password, keep the
   `Verify Remote Cert` checkbox enabled, paste the matching CA certificate
   into the `CA Certificate` field, click test connection and save the
   endpoint.
3. In `Administration->Registries` page, edit the endpoint and replace the CA
   certificate with one that does not match the remote endpoint's
   certificate, keeping `Verify Remote Cert` enabled, then click test
   connection.

# Expected Outcome:

* In step 2, test connection succeeds and the endpoint is saved, since the
  supplied CA certificate matches the remote endpoint's certificate.
* In step 3, test connection fails, since the CA certificate no longer
  matches the remote endpoint's certificate.

# Possible Problems:
None
