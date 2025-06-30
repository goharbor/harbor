## Harbor Roadmap

### About this document

This document provides a link to the [Harbor Project board](https://github.com/orgs/goharbor/projects/1) that serves as the up-to-date description of items that are in the Harbor release pipeline. The board has separate swim lanes for each release. Most items are gathered from the community or include a feedback loop with the community. This should serve as a reference point for Harbor users and contributors to understand where the project is heading, and help determine if a contribution could be conflicting with a longer term plan.

### How to help?

Discussion on the roadmap can take place in threads under [Issues](https://github.com/goharbor/harbor/issues) or in [community meetings](https://goharbor.io/community/). Please open and comment on an issue if you want to provide suggestions and feedback to an item in the roadmap. Please review the roadmap to avoid potential duplicated effort.

### How to add an item to the roadmap?
Please open an issue to track any initiative on the roadmap of Harbor (Usually driven by new feature requests). We will work with and rely on our community to focus our efforts on improving Harbor.

### Current Roadmap

The following table includes the current roadmap for Harbor. If you have any questions or would like to contribute to Harbor, please attend a [community meeting](https://goharbor.io/community/) to discuss with our team. If you don't know where to start, we are always looking for contributors who will help us reduce technical, automation, and documentation debt. Please take the timelines & dates as proposals and goals. Priorities and requirements change based on community feedback, roadblocks encountered, community contributions, etc. If you depend on a specific item, we encourage you to attend community meetings to get updated status information, or help us deliver that feature by contributing to Harbor.


`Last Updated: June 2022`

|Theme|Description|Timeline|
|--|--|--|
|Harbor for Edge|Optimize data transfer for substandard data center connectivity to edge nodes|2022 H2|
||Create a lightweight Harbor with reduced feature set for minimal footprint|2022 H2|
|Deployment|Improve Kubernetes Operator for Harbor, enabling improved Day1 and Day2 operations including enterprise grade HA, faster deployments and upgrades, automate backups and restores, and sensible defaults|Future|
||Support Notary v2 to deliver persisting image signatures across image replications|Future|
|Optimized Scalability & Performance|Introduce cache layer to improve performance|2022 H2|
|Image acceleration|Leverage Nydus to support image acceleratiion|Future|
|Regex Support|Add full Regex support to all modules within Harbor consistently including configuration of replication policies, retention policies, immutability policies and more|Future|
|ARM Harbor|release an ARM deployment of Harbor|Future|
|Backup & Restore|Leverage Project Velero to offer application-aware Harbor backup, including databases, Kubernetes objects and Persistent Volumes|2022 H2|
|Extended image support|Support WASM images|2022 H2|
||System artifact manager|2022 H2|
|CVE reporting|Export CVE list at the repo level|Future|
|SBoM support|SBoM generation & attestation|Future|
|Networking|Support dual stack IPv6/IPv4 network for Harbor pods in a Kubernetes cluster|2022 H2|


### Completed Items

|Theme|Description|Timeline|
|--|--|--|
|Security Analysis|Leverage Sysdig Secure scanner to analyze container images|Aug 2020|
|Image Distribution|Ability for Harbor registry to function as a pull-through cache for remote registry|Sep 2020|
|Performance & Reliability|Non-blocking Garbage Collection|Sep 2020|
|Image Distribution|Leverage P2P providers like Alibaba Dragonfly and Uber Kraken to geo-distribute artifacts at higher rates|Oct 2020|
|Extensibility|Allow vendors to publish and share OCI artifacts like Machine Learning (Kubeflow) workloads generated datatypes on Harbor|Oct 2020|
|Registry|Improve support for Windows containers layers|Oct 2020|
|I&AM and RBAC|Improved Multi-tenancy through granular access and ability to manage teams of users and robot accounts through workspaces|Dec 2020|
|Observability|Expose Harbor metrics through Prometheus Integration|Mar 2021|
|Tracing|Leverage OpenTelemetry for enhanced tracing capabilities and identify bottlenecks and improve performance |Mar 2021|
|Image Signing|Leverage Sigstore Cosign to deliver persistent image signatures across image replications|Apr 2021|
