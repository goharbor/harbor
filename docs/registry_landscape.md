# Registry Landscape  
The cloud native ecosystem is moving rapidly–registries and their featuresets are no exception. We've made our best effort to survey the container registry landscape and compare to our core featureset.

If you find something outdated or outright erroneous, please submit a PR and we'll fix it right away.

| Feature                                                | Harbor | Docker Trusted Registry | Quay | Cloud Providers (GCP, AWS, Azure) | Docker Distribution         | Artifactory |
| -------------:                                         | :----: | :---------------------: | :--: | :-------------------------------: | :-----------------:         | :---------: |
| Single Sign-On                                         | ✗      | ✓                       | ?    | ✓                                 | ✗                           | ✓           |
| Local Auth                                             | ✓      | ✓                       | ✓    | ✓                                 | ✗                           | ✓           |
| LDAP-based Auth                                        | ✓      | ✓                       | ✓    | partial                           | ✗                           | ✓           |
| Audit Logs                                             | ✓      | ✓                       | ✓    | ✓                                 | ✗                           | ✓           |
| Metadata (registry configuration) Replication          | ✗      | ✓                       | ✓    | n/a                               | ✗                           | ✓           |
| CI Integration / Build from Dockerfile                 | ✗      | ✓                       | ✓    | requires additional tooling       | requires additional tooling | ✓           |
| See what lines were used to produce image              | ✗      | ✓                       | ?    | ✗                                 | ✗                           | ✓           |
| Upstream Registry Proxy Cache                          | ✗      | ✓                       | ✗    | ✗                                 | ✓                           | ✓           |
| Content Trust and Validation                           | ✓      | ✓                       | ✗    | ✗                                 | partial                     | partial     |
| Vulnerability Scanning & Monitoring                    | ✓      | ✓                       | ✓    | ✗                                 | ✗                           | ✓           |
| Replication                                            | ✓      | ✓                       | ✓    | n/a                               | ✗                           | ✓           |
| Multi-Tenancy (projects, teams, etc.)                  | ✓      | ✓                       | ✓    | partial                           | ✗                           | ✓           |
| Tag Immutability Support                               | ✗      | ✓                       | ✗    | ✗                                 | ✗                           | ?           |
| Role-Based Access Control                              | ✓      | ✓                       | ✓    | ✓                                 | ✗                           | ✓           |
| Custom TLS Certificates                                | ✓      | ✓                       | ✓    | ✗                                 | ✓                           | ✓           |
| Ability to Determine Version of Binaries in Containers | ✓      | ✓                       | ✓    | ✗                                 | ✗                           | ?           |


