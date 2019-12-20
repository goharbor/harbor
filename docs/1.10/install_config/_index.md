[Back to table of contents](../_index.md)

----------

# Harbor Installation and Configuration

This section describes how to perform a new installation of Harbor.

If you are upgrading from a previous version of Harbor, you might need to update the configuration file and migrate your data to fit the database schema of the later version. For information about upgrading, see [Upgrading Harbor](../../administration/upgrade/_index.md).

You can also use Helm to install Harbor on a Kubernetes cluster, to make it highly available. For information about installing Harbor with Helm on a Kubernetes cluster, see the [Harbor High Availability Guide](https://github.com/goharbor/harbor-helm/blob/master/docs/High%20Availability.md) in the https://github.com/goharbor/harbor-helm repository.

Before you install Harbor, you can test its functionality on a demo server that the Harbor team has made available. For information, see [Test Harbor with the Demo Server](demo_server.md).

You can use Harbor with different 3rd party replication adapters, OIDC adapters, and scanner adapters. For information about the supported adapters, see the [Harbor Compatibility List](harbor_compatibility_list.md).

## Installation Process

The standard Harbor installation process involves the following stages:

1. Make sure that your target host meets the [Harbor Installation Prerequisites](installation_prereqs.md).
1. [Download the Harbor Installer](download_installer.md)
1. [Configure HTTPS Access to Harbor](configure_https.md)
1. [Configure the Harbor YML File](configure_yml_file.md)
1. [Run the Installer Script](run_installer_script.md)

If installation fails, see [Troubleshooting Harbor Installation
](troubleshoot_installation.md).

## Post-Installation Configuration

For information about how manage your deployed Harbor instance, see [Reconfigure Harbor and Manage the Harbor Lifecycle](reconfigure_manage_lifecycle.md). 

By default, Harbor uses its own private key and certificate to authenticate with Docker. For information about how to optionally customize your configuration to use your own key and certificate, see [Customize the Harbor Token Service](customize_token_service.md).

After installation, you perform configuration operations in the Harbor interface. However, Harbor also provides a command line interface (CLI) that allows yoy to [Configure Harbor User Settings at the Command Line](configure_user_settings_cli.md).

## Harbor Components

The table below lists the components that are deployed when you deploy Harbor.

|Component|Version|
|---|---|
|Postgresql|9.6.10-1.ph2|
|Redis|4.0.10-1.ph2|
|Clair|2.0.8|
|Beego|1.9.0|
|Chartmuseum|0.9.0|
|Docker/distribution|2.7.1|
|Docker/notary|0.6.1|
|Helm|2.9.1|
|Swagger-ui|3.22.1|

----------
[Back to table of contents](../../_index.md)