## Harbor Roadmap

### About this document

This document provides a link to the [Harbor Project board](https://github.com/orgs/goharbor/projects/1) that serves as the up to date description of items that are in the Harbor release pipeline. The board has separate swim lanes for each release. Most items are gathered from the community or include a feedback loop with the community. This should serve as a reference point for Harbor users and contributors to understand where the project is heading, and help determine if a contribution could be conflicting with a longer term plan.

### How to help?

Discussion on the roadmap can take place in threads under [Issues](https://github.com/goharbor/harbor/issues) or in [community meetings](https://github.com/goharbor/community/blob/master/MEETING_SCHEDULE.md). Please open and comment on an issue if you want to provide suggestions and feedback to an item in the roadmap. Please review the roadmap to avoid potential duplicated effort.

### How to add an item to the roadmap?
Please open an issue to track any initiative on the roadmap of Harbor (Usually driven by new feature requests). We will work with and rely on our community to focus our efforts to improve Harbor.

### Current Roadmap

The following table includes the current roadmap for Harbor. If you have any questions or would like to contribute to Harbor, please attend a [community meeting](https://github.com/goharbor/community/wiki/Harbor-Community-Meetings) to discuss with our team. If you don't know where to start, we are always looking for contributors that will help us reduce technical, automation, and documentation debt. Please take the timelines & dates as proposals and goals. Priorities and requirements change based on community feedback, roadblocks encountered, community contributions, etc. If you depend on a specific item, we encourage you to attend community meetings to get updated status information, or help us deliver that feature by contributing to Harbor.


`Last Updated: June 2020`

|Theme|Description|Timeline|
|--|--|--|
|Proxy Cache|Ability for Harbor registry to function as a pull-through cache for remote registry|Aug 2020|
|Non-blocking Garbage Collection|GC to run silent with no impact to artifact push / pull / deletion|Aug 2020|
|P2P Integration|Leverage P2P providers like Alibaba Dragonfly and Uber Kraken to geo-distribute artifacts at higher rates|August 2020|
|Sysdig Harbor Scanner|Leverage Sysdig Secure scanner to analyze container images|Aug 2020|
|HA (high availability)|A Harbor k8s operator for better Day1 and Day2 operations including enterprise grade HA|Aug 2020|
|OCI Artifact Type Extender|Allows vendors to publish and share OCI artifacts like Machine Learning (Kubeflow) workloads generated datatypes on Harbor|Aug 2020|
|System-lvl Robot Accounts|Service accounts with proper RBAC and differentiated access permissions|Dec 2020|
|Improved Multi-tenancy|Granulated access and ability to manage teams of users and robot accounts through workspaces|Dec 2020|
|Windows Containers|Harbor to support pushing, pulling [Windows Containers running in Kubernetes](https://docs.microsoft.com/en-us/virtualization/windowscontainers/kubernetes/getting-started-kubernetes-windows)|Dec 2020|
|Prometheus Integration|Expose Harbor metrics by publishing over Prometheus|early 2021|
|Backup & Restore|Leverage Velero to backup Harbor including databases, Kubernetes objects and persistent volumes|early 2021|
|IPv6 support|Harbor running in a IPv6-only network on Kubernetes clusters|2021|
