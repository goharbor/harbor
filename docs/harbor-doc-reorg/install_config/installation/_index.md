# Installing Harbor

The Harbor installation process involves the following stages:

1. Make sure that your target host meets the [Harbor Installation Prerequisites](installation_prereqs.md).
1. [Download the Harbor Installer](download_installer.md)
1. [Configure the Harbor YML File](configure_yml_file.md)
1. [Run the Installer Script](run_installer_script.md)

Harbor does not ship with any certificates, and, by default, uses HTTP to serve requests. While this makes it relatively simple to set up and run - especially for a development or testing environment - it is **not** recommended for a production environment.  To enable HTTPS, see [Configure HTTPS Access to Harbor](../configuration/configure_https.md).

**NOTE**: If you run a previous version of Harbor, you may need to update ```harbor.yml``` and migrate the data to fit the new database schema. For more details, please refer to **[Harbor Migration Guide](migration_guide.md)**.

In addition, the deployment instructions on Kubernetes has been created by the community. Refer to [Harbor on Kubernetes](kubernetes_deployment.md) for details.

## Harbor Components

The table below lists the components that are deployed when you install this version of Harbor.

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