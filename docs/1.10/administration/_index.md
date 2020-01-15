---
title: Harbor Administration
---

This section describes how to configure and maintain Harbor after deployment. These operations are performed by the Harbor system administrator. The Harbor system administrator performs global configuration operations that apply to the whole Harbor instance.

The operations that are performed by the Harbor system administrator are the following.

- Select database, LDAP/Active Directory, or OIDC based authentication. For information, see [Configuring Authentication](configure-authentication/_index.md).
- Add users in database authentication mode and assign the system administrator role to other users. For information, see [Role Based Access Control](managing-users/rbac.md).
- Configure general system settings, including setting up an email server and setting the registry to read-only mode. For information, see [Configure Global Settings](general-settings.md).
- Configure how projects are created, and apply resource quotas to projects. For information, see [Configure Project Settings](configure-project-settings.md).
- Set up replication of images between Harbor and another Harbor instance or a 3rd party replication target. For information, see [Configuring Replication](configuring-replication/_index.md).
- Set up vulnerability scanners to check the images in the registry for CVE vulnerabilities. For information, see [Vulnerability Scanning](vulnerability-scanning/_index_.md).
- Perform garbage collection, to remove unnecessary data from Harbor. For information, see [Garbage Collection](garbage-collection.md).
- Upgrade Harbor when a new version becomes available. For information, see [Upgrading Harbor](upgrade/_index_.md).
