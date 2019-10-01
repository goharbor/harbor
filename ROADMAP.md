## Harbor Roadmap

### About this document

This document provides description of items that are gathered from the community and planned in Harbor's roadmap. This should serve as a reference point for Harbor users and contributors to understand where the project is heading, and help determine if a contribution could be conflicting with a longer term plan.

### How to help?

Discussion on the roadmap can take place in threads under [Issues](https://github.com/vmware/harbor/issues). Please open and comment on an issue if you want to provide suggestions and feedback to an item in the roadmap. Please review the roadmap to avoid potential duplicated effort.

### How to add an item to the roadmap?
Please open an issue to track any initiative on the roadmap of Harbor. We will work with and rely on our community to focus our efforts to improve Harbor.


---

### 1. Notary
The notary feature allows publishers to sign their images offline and to push the signed content to a notary server. This ensures the authenticity of images.

### 2. Vulnerability Scanning
The capability to scan images for vulnerability.

### 3. Image replication enhancement
To provide more sophisticated rule for image replication.
- Image filtering by tags
- Replication can be scheduled at a certain time using a rule like: one time only, daily, weekly, etc.
- Image deletion can have the option not to be replicated to a remote instance.
- Global replication rule: Instead of setting the rule of individual project, system admin can set a global rule for all projects.
- Project admin can set replication policy of the project.

### 4. Authentication (OAuth2)
In addition to LDAP/AD and local users, OAuth 2.0 can be used to authenticate a user.

### 5. High Availability
Support multi-node deployment of Harbor for high availability, scalability and load-balancing purposes.

### 6. Statistics and description for repositories
User can add a description to a repository. The access count of a repo can be aggregated and displayed.

### 7. Migration tool to move from an existing registry to Harbor
A tool to migrate images from a vanilla registry server to Harbor, without the need to export/import a large amount of data.
