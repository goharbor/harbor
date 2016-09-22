# Changelog

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
