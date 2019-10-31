# Installing Harbor

This guide describes how to install and configure Harbor for the first time.

If you are upgrading from a previous version of Harbor, you might need to update the configuration file and migrate your data to fit the database schema of the later version. For information about upgrading, see the [Upgrading Harbor](../../administration/upgrade/_index.md).

You can also use Helm to install Harbor on a Kubernetes cluster, to make it highly available. For information about installing Harbor with Helm on a Kubernetes cluster, see the [Harbor High Availability Guide](https://github.com/goharbor/harbor-helm/blob/master/docs/High%20Availability.md) in the https://github.com/goharbor/harbor-helm repository.

Before you install Harbor, you can test its functionality on a demo server that the Harbor team has made available. For information, see [Test Harbor with the Demo Server](demo_server.md).

The standard Harbor installation process involves the following stages:

1. Make sure that your target host meets the [Harbor Installation Prerequisites](installation_prereqs.md).
1. [Configure HTTPS Access to Harbor](configure_https.md)
1. [Download the Harbor Installer](download_installer.md)
1. [Configure the Harbor YML File](configure_yml_file.md)
1. [Run the Installer Script](run_installer_script.md)

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