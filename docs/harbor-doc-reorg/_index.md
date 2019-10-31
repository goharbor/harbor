# Harbor Documentation 

This is the main table of contents for the Harbor documentation.

## Harbor Installation and Configuration

This section describes how to install Harbor and perform the required initial configurations. These day 1 operations are performed by the Harbor Administrator.

- [Installing Harbor](install_config/installation/_index.md)
  - [Test Harbor with the Demo Server](install_config/installation/demo_server.md)
  - [Harbor Installation Prerequisites](install_config/installation/installation_prereqs.md)
  - [Configure HTTPS Access to Harbor](install_config/installation/configure_https.md)
  - [Download the Harbor Installer](install_config/installation/download_installer.md)
  - [Configure the Harbor YML File](install_config/installation/configure_yml_file.md)
  - [Run the Installer Script](install_config/installation/run_installer_script.md)
  - [Troubleshooting Harbor Installation
](install_config/installation/troubleshoot_installation.md)
- [Post-Installation Configuration](install_config/configuration/_index.md)
  - [Reconfigure Harbor and Manage the Harbor Lifecycle](install_config/configuration/reconfigure_manage_lifecycle.md)
  - [Customize the Harbor Token Service](install_config/configuration/customize_token_service.md)
  - [Configure Notary Content Trust](install_config/configuration/configure_notary_content_trust.md)
- [Initial Configuration in the Harbor UI](install_config/configuration/initial_config_ui.md)
  - [Configure Authentication](install_config/configuration/configure_authentication.md)
  - [Other General Settings](install_config/configuration/general_settings.md)
  
You can also use Helm to install Harbor on a Kubernetes cluster, to make it highly available. For information about installing Harbor with Helm on a Kubernetes cluster, see the [Harbor High Availability Guide](https://github.com/goharbor/harbor-helm/blob/master/docs/High%20Availability.md) in the https://github.com/goharbor/harbor-helm repository.

## Harbor Administration

This section describes how to use and maintain Harbor after deployment. These day 2 operations are performed by the Harbor Administrator.

- [Managing Users](administration/managing_users/_index.md)
  - [Harbor Role Based Access Control (RBAC)](administration/managing_users/configure_rbac.md)
  - [User Permissions By Role](administration/managing_users/user_permissions_by_role.md)
  - [Configure Harbor User Settings at the Command Line](administration/managing_users/configure_user_settings_cli.md)
  - [Manage Roles by LDAP Group](administration/managing_users/manage_role_by_ldap_group.md)
- [Configure Project Settings](administration/configure_project_settings/_index.md)
  - [Set Project Quotas](administration/configure_project_settings/set_project_quotas.md)
- [Configuring Replication](administration/configuring_replication/_index.md)
  - [Create Replication Endpoints](administration/configuring_replication/create_replication_endpoints.md)
  - [Create Replication Rules](administration/configuring_replication/create_replication_rules.md)
  - [Manage Replications](administration/configuring_replication/manage_replications.md) 
- [Vulnerability Scanning with Clair](administration/vulnerability_scanning/_index.md)
  - [Scan an Individual Image](administration/vulnerability_scanning/scan_individual_image.md)
  - [Scan All Images](administration/vulnerability_scanning/scan_all_images.md)
  - [Schedule Scans](administration/vulnerability_scanning/schedule_scans.md)
  - [Import Vulnerability Data to an Offline Harbor instance](administration/vulnerability_scanning/import_vulnerability_data.md)
  - [Configure System-Wide CVE Whitelists](administration/vulnerability_scanning/configire_system_whitelist.md)
- [Garbage Collection](administration/garbage_collection/_index.md)
- [Upgrading Harbor](administration/upgrade/_index.md)
  - [Upgrade Harbor and Migrate Data](administration/upgrade/upgrade_migrate_data.md)
  - [Roll Back an Upgrade](administration/upgrade/roll_back_upgrade.md)
- [Manage the Harbor Instance](administration/manage_harbor/_index.md)
  - 
  - [Access Harbor Logs](administration/manage_harbor/access_logs.md)

## Working with Harbor Projects

This section describes how users with the developer, master, and project administrator roles manage and participate in Harbor projects.

- [Configure a Per-Project CVE Whitelist](working_with_projects/configure_project_whitelist.md)
- [](working_with_projects/)
- [](working_with_projects/)
- [](working_with_projects/)
- [](working_with_projects/)
- [](working_with_projects/)

## Build, Customize, and Contribute to Harbor

This section describes how developers can build from Harbor source code, customize their deployments, and contribute to the open-source Harbor project.

- [](build_customize_contribute/)
- [](build_customize_contribute/)
- [](build_customize_contribute/)
- [](build_customize_contribute/)
- [](build_customize_contribute/)
- [](build_customize_contribute/)