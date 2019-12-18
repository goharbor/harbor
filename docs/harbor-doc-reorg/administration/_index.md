# Harbor Administration

This section describes how to configure and maintain Harbor after deployment. These operations are performed by the Harbor system administrator. The Harbor system administrator performs global configuration operations that apply to the whole Harbor instance.

The operations that are performed by the Harbor system administrator are the following.

- Select database, LDAP/Active Directory, or OIDC based authentication. For information, see [Configuring Authentication](configure_authentication/configure_authentication.md).
- Add users in database authentication mode and assign the system administrator role to other users. For information, see [Role Based Access Control](managing_users/rbac.md).
- Configure general system settings, including setting up an email server and setting the registry to read-only mode. For information, see [Configure Global Settings](general_settings.md).
- Configure how projects are created, and apply resource quotas to projects. For information, see [Configure Project Settings](configure_project_settings.md)
- Set up replication of images between Harbor and another Harbor instance or a 3rd party replication target. For information, see [Configuring Replication](configuring_replication/configuring_replication.md).
- Set up vulnerability scanners to check the images in the registry for CVE vulnerabilities. For information, see [Vulnerability Scanning](vulnerability_scanning/vulnerability_scanning.md).
- Perform garbage collection, to remove unnecessary data from Harbor. For information, see [Garbage Collection](garbage_collection.md).
- Upgrade Harbor when a new version becomes available. For information, see [Upgrading Harbor](upgrade/upgrade_harbor.md).

----------

[Back to table of contents](../_index.md)