# Installing Harbor

This guide describes how to install and configure Harbor by using either the online or offline installer. The installation processes are almost the same.

If you are upgrading from a previous version of Harbor, you might need to update the configuration file and migrate your data to fit the database schema of the later version. For information about upgrading, see the **[Harbor Upgrade and Migration Guide](migration_guide.md)**.

In addition, the Harbor community created instructions describing how to deploy Harbor on Kubernetes. If you want to deploy Harbor to Kubernetes, see [Harbor on Kubernetes](kubernetes_deployment.md).

The Harbor installation process involves the following stages:

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