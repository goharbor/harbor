# Changelog

# v1.8.0 (2019-05-21)
[Full list of issues fixed in v1.8.0](https://github.com/goharbor/harbor/issues?q=is%3Aissue+is%3Aclosed+label%3Atarget%2F1.8.0)
* Support for OpenID Connect - OpenID Connect (OIDC) is an authentication layer on top of OAuth 2.0, allowing Harbor to verify the identity of users based on the authentication performed by an external authorization server or identity provider.
* Robot accounts - Robot accounts can be configured to provide administrators with a token that can be granted appropriate permissions for pulling or pushing images. Harbor users can continue operating Harbor using their enterprise SSO credentials, and use robot accounts for CI/CD systems that perform Docker client commands.
* Replication advancements - Harbor new version replication allows you to replicate your Harbor repository to and from non-Harbor registries. Harbor 1.8 expands on the Harbor-to-Harbor replication feature, adding the ability to replicate resources between Harbor and Docker Hub, Docker Registry, and Huawei Registry. This is enabled through both push and pull mode replication.
* Health check API, showing detailed status and health of all Harbor components.
* Support for defining cron-based scheduled tasks in the Harbor UI. Administrators can now use cron strings to define the schedule of a job. Scan, garbage collection and replication jobs are all supported.
API explorer integration. End users can now explore and trigger Harbor’s API via the swagger UI nested inside Harbor’s UI.
* Introduce a new master role to project, the role's permissions are more than developer and less than project admin.
* Introduce harbor.yml as the replacement of harbor.cfg and refactor the prepare script to provide more flexibility to the installation process based on docker-compose
* Enhancement of the Job Service engine to include webhook events, additional APIs for automation, and numerous bug fixes to improve the stability of the service.
* Docker Registry upgraded to v2.7.1.

## v1.7.5 (2019-04-02)
* Bumped up Clair to v2.0.8
* Fixed issues in supporting windows images. #6992 #6369
* Removed user-agent check-in notification handler. #5729
* Fixed the issue global search not working if chartmusuem is not installed #6753

## v1.7.4 (2019-03-04)
[Full list of issues fixed in v1.7.4](https://github.com/goharbor/harbor/issues?q=is%3Aissue+is%3Aclosed+label%3Atarget%2F1.7.4)

## v1.7.1 (2019-01-07)
[Full list of issues fixed in v1.7.1](https://github.com/goharbor/harbor/issues?q=is%3Aissue+is%3Aclosed+label%3Atarget%2F1.7.1)

## v1.7.0 (2018-12-19)
* Support deploy Harbor with Helm Chart, enables the user to have high availability of Harbor services, refer to the [Installation and Configuration Guide](https://github.com/goharbor/harbor-helm/tree/1.0.0). 
* Support on-demand Garbage Collection, enables the admin to configure run docker registry garbage collection manually or automatically with a cron schedule.
* Support Image Retag, enables the user to tag image to different repositories and projects, this is particularly useful in cases when images need to be retagged programmatically in a CI pipeline.
* Support Image Build History, makes it easy to see the contents of a container image, refer to the [User Guide](https://github.com/goharbor/harbor/blob/release-1.7.0/docs/user_guide.md#build-history).
* Support Logger customization, enables the user to customize STDOUT / STDERR / FILE / DB logger of running jobs.
* Improve user experience of Helm Chart Repository:
   - Chart searching included in the global search results
   - Show chart versions total number in the chart list
   - Mark labels to helm charts
   - The latest version can be downloaded as default one on the chart list view
   - The chart can be deleted by deleting all the versions under it


## v1.6.0 (2018-09-11)

- Support manages Helm Charts: From version 1.6.0, Harbor is upgraded to be a composite cloud-native registry, which supports both image management and helm charts management.
- Support LDAP group: User can import an LDAP/AD group to Harbor and assign project roles to it.
- Replicate images with label filter: Use newly added label filter to narrow down the sourcing image list when doing image replication.
- Migrate multiple databases to one unified PostgreSQL database.

## v1.5.0 (2018-05-07)

- Support read-only mode for registry: Admin can set registry to read-only mode before GC. [Details](https://github.com/vmware/harbor/blob/master/docs/user_guide.md#managing-registry-read-only)
- Label support: User can add label to image/repository, and filter images by label on UI/API. [Details](https://github.com/vmware/harbor/blob/master/docs/user_guide.md#managing-labels)
- Show repositories via Cardview.
- Re-work Job service to make it HA ready.

## v1.4.0 (2018-02-07)

- Replication policy rework to support wildcard, scheduled replication.
- Support repository level description.
- Batch operation on projects/repositories/users from UI.
- On board LDAP user when adding member to a project.

## v1.3.0 (2018-01-04)

- Project level policies for blocking the pull of images with vulnerabilities and unknown provenance.
- Remote certificate verification of replication moved to target level.
- Refined all images to improve security.

## v1.2.0 (2017-09-15)

- Authentication and authorization, implementing vCenter Single Sign On across components and role-based access control at the project level. [Read more](https://vmware.github.io/vic-product/assets/files/html/1.2/vic_overview/introduction.html#projects)
- Full integration of the vSphere Integrated Containers Registry and Management Portal user interfaces. [Read more](https://vmware.github.io/vic-product/assets/files/html/1.2/vic_cloud_admin/)
- Image vulnerabilities scanning.

## v1.1.0 (2017-04-18)

- Add in Notary support
- User can update configuration through Harbor UI
- Redesign of Harbor's UI using Clarity
- Some changes to API
- Fix some security issues in token service
- Upgrade base image of nginx for latest openssl version
- Various bug fixes.

## v0.5.0 (2016-12-6)

- Refactory for a new build process
- Easier configuration for HTTPS in prepare script
- Script to collect logs of a Harbor deployment
- User can view the storage usage (default location) of Harbor.
- Add an attribute to disable normal user to create project
- Various bug fixes.

For Harbor virtual appliance:

- Improve the bootstrap process of ova installation.
- Enable HTTPS by default for .ova deployment, users can download the default root cert from UI for docker client or VCH.
- Preload a photon:1.0 image to Harbor for users who have no internet connection.

## v0.4.5 (2016-10-31)

- Virtual appliance of Harbor for vSphere.
- Refactory for new build process.
- Easier configuration for HTTPS in prepare step.
- Updated documents.
- Various bug fixes.

## v0.4.0 (2016-09-23)

- Database schema changed, data migration/upgrade is needed for previous version.
- A project can be deleted when no images and policies are under it.
- Deleted users can be recreated.
- Replication policy can be deleted.
- Enhanced LDAP authentication, allowing multiple uid attributes.
- Pagination in UI.
- Improved authentication for remote image replication.
- Display release version in UI
- Offline installer.
- Various bug fixes.

## v0.3.5 (2016-08-13)

- Vendoring all dependencies and remove go get from dockerfile
- Installer using Docker Hub to download images
- Harbor base images moved to Photon OS (except for official images from third party)
- New Harbor logo
- Various bug fixes

## v0.3.0 (2016-07-15)

- Database schema changed, data migration/upgrade is needed for previous version.
- New UI
- Image replication across multiple registry instances
- Integration with registry v2.4.0 to support image deletion and garbage collection
- Database migration tool
- Bug fixes

## v0.1.1 (2016-04-08)

- Refactored database schema
- Migrate to docker-compose v2 template
- Update token service to support layer mount
- Various bug fixes

## v0.1.0 (2016-03-11)

Initial release, key features include

- Role based access control (RBAC)
- LDAP / AD integration
- Graphical user interface (GUI)
- Auditting and logging
- RESTful API
- Internationalization
