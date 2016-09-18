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

### 2. Image replication between Harbor instances (Completed)
Enable images to be replicated between two or more Harbor instances. This is useful to have multiple registry servers servicing a large cluster of nodes, or have distributed registry instances with identical images.

### 3. Image deletion and garbage collection (Completed)
a) Images can be deleted from UI. The files of deleted images are not removed immediately. 

b) The files of deleted images are recycled by an administrator during system maintenance(Garbage collection). The registry service must be shut down during the process of garbage collection.


### 4. Authentication (OAuth2) 
In addition to LDAP/AD and local users, OAuth 2.0 can be used to authenticate a user.

### 5. High Availability 
Support multi-node deployment of Harbor for high availability, scalability and load-balancing purposes.

### 6. Statistics and description for repositories
User can add a description to a repository. The access count of a repo can be aggregated and displayed.


### 7. Audit all operations in the system
Currently only image related operations are logged. Other operations in Harbor, such as user creation/deletion, role changes, password reset, should be tracked as well.


### 8. Migration tool to move from an existing registry to Harbor 
A tool to migrate images from a vanilla registry server to Harbor, without the need to export/import a large amount of data.
