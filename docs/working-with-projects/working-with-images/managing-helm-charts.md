---
title: Managing Helm Charts
weight: 95
---

[Helm](https://helm.sh) is a package manager for [Kubernetes](https://kubernetes.io). Helm uses a packaging format called [charts](https://helm.sh/docs/topics/charts). Since version 1.6.0, Harbor is a composite, cloud-native registry which supports both container image management and Helm charts management. Access to Helm charts in Harbor is controlled by [role-based access controls (RBAC)](https://en.wikipedia.org/wiki/Role-based_access_control) and is restricted by projects.

There are two places to manage Helm charts.
- The ChartMuseum, provided by Harbor since version 1.6.0
- The OCI-compatible registry, provided by Harbor since version 2.0.0  
This means you can manage Helm charts alongside your container images through the same set of projects and repositories.

## Manage Helm Charts with the ChartMuseum in Harbor Interface

### List charts

Log in to the Harbor interface, click **Projects**, and select your project to go to the project detail page. Click the **Helm Charts** tab to view the existing Helm charts with the following information:
* Name of Helm chart
* The status of the chart: Active or Deprecated
* The number of chart versions
* The time the chart was created

![list charts](../../../img/list-charts.png)

You can click the icons on the top right to switch between card view and list view.

### Upload a New Chart

1. Click **UPLOAD** to open the chart uploading dialog. 
1. Choose the chart to upload from your filesystem. 
1. Click **UPLOAD** to upload it to the chart repository server.

![upload charts](../../../img/upload-charts.png)

If the chart is signed, you can choose the corresponding provenance file from your filesystem, and then click **UPLOAD** to upload them together at the same time.

If the chart uploads successfully, it is displayed in the chart list.

### List Chart Versions

To see the available versions of a chart, click the chart name in the list of charts. The chart detail shows the following information: 
* the chart version number
* the maintainers of the chart version
* the template engine used (default is `gotpl`)
* the created timestamp of the chart version

![list charts versions](../../../img/list-chart-versions.png)

There is at least one version for each chart in the chart list. You can click the icons on the top right to switch between card view and list view.

Click the checkbox in the first column to select one or more chart versions, and then you can perform the following actions:
* Click **DELETE** to delete the selected chart versions from the chart repository server. Batch operation is supported.
* Click **DOWNLOAD** to download the chart artifact file. Batch operation is not supported.
* Click **UPLOAD** to upload a new version for the current chart.

### Adding Labels to and Removing Labels from Chart Versions
If you have Harbor system administrator, project administrator, or project developer role, you can click **ADD LABELS** to add labels to or remove labels from chart versions.

![add labels to chart versions](../../../img/add-labels-to-chart-versions.png)


### Filtering Chart Versions by Label
The chart versions can be filtered by labels:

![filter chart versions by labels](../../../img/filter-chart-versions-by-label.png)

### View Chart Version Details
Click the chart version number link to open the chart version details view. You can see more details about the specified chart version here. There are three content sections:
* **Summary:**
  * readme of the chart
  * overall metadata like home, created timestamp and application version
  * related Helm commands for reference, such as `helm add repo` and `helm install`.
![chart details](../../../img/chart-details.png)
* **Dependencies:**
  * shows the name, version, and repository of dependent sun charts
![chart dependencies](../../../img/chart-dependencies.png)
* **Values:**
  * shows the content from the `values.yaml` file with highlight code preview
  * You can click the icon buttons on the top right to toggle between the YAML file view and key-value pair list view.
![chart values](../../../img/chart-values.png)

You can click the **DOWNLOAD** button on the top right to download the YAML file.

## Working with ChartMuseum via the Helm CLI

As a Helm chart repository, Harbor can interoperate with the Helm CLI. To install the Helm CLI, see [Installing Helm](https://helm.sh/docs/intro/install/). Run the command `helm version` to make sure the version of Helm CLI is v2.9.1 or greater.

```sh
helm version

#Client: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}
#Server: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}
```

### Add Harbor to the Repository List

Before working with the Helm CLI, add Harbor to the repository list with `helm repo add` command. Two different modes are supported.

* Add Harbor as a unified single index entry point.

    In this mode Helm can be made aware of all the charts located in different projects and which are accessible by the currently authenticated user.

    ```sh
    helm repo add --ca-file ca.crt --username=admin --password=Passw0rd myrepo https://xx.xx.xx.xx/chartrepo
    ```

   {{< note >}}
   Provide both a CA file and cert files. This is necessary due to an issue in Helm.
   {{< /note >}}

* Add Harbor project as separate index entry point.

    In this mode, Helm can only pull charts in the specified project.

    ```sh
    helm repo add --ca-file ca.crt --username=admin --password=Passw0rd myrepo https://xx.xx.xx.xx/chartrepo/myproject
    ```

### Push Charts to the Repository Server with the CLI

As an alternative, you can also upload charts using the CLI. It is not supported by the native Helm CLI. A plugin from the community must be installed before pushing. Run the following `helm plugin install` command to install the `helm-push` plugin first.

```sh
helm plugin install https://github.com/chartmuseum/helm-push
```

After a successful installation, run the `push` command to upload your charts:

```sh
helm push --ca-file=ca.crt --username=admin --password=passw0rd chart_repo/hello-helm-0.1.0.tgz myrepo
```

{{< note >}}
The `push` command does not yet support pushing a prov file of a signed chart.
{{< /note >}}

### Install Charts

Before installing, make sure Helm is correctly initialized with the `helm init` command, and the chart index is synchronized with the `helm repo update` command.

Search the chart with the keyword if you're not sure where it is:

```sh
helm search hello

#NAME                            CHART VERSION   APP VERSION     DESCRIPTION
#local/hello-helm                0.3.10          1.3             A Helm chart for Kubernetes
#myrepo/chart_repo/hello-helm    0.1.10          1.2             A Helm chart for Kubernetes
#myrepo/library/hello-helm       0.3.10          1.3             A Helm chart for Kubernetes
```

If everything is ready, install the chart in your Kubernetes cluster:

```sh
helm install --ca-file=ca.crt --username=admin --password=Passw0rd --version 0.1.10 repo248/chart_repo/hello-helm
```

For other more Helm commands like how to sign a chart, please refer to the [Helm documentation](https://docs.helm.sh/helm/#helm).


## Manage Helm Charts with the OCI-Compatible Registry of Harbor

Helm 3 supports registry operations for an OCI-compatible registry including pushing and pulling. To install the latest Helm CLI, see [Installing Helm](https://helm.sh/docs/intro/install/). Also, run `helm version` to make sure the version of Helm CLI is v3.0.0+.

```sh
helm version

#version.BuildInfo{Version:"v3.2.1", GitCommit:"fe51cd1e31e6a202cba7dead9552a6d418ded79a", GitTreeState:"clean", GoVersion:"go1.13.10"}
```

### Login to the OCI-Compatible Registry of Harbor

Before pulling or pushing Helm charts with the OCI-compatible registry of Harbor, Harbor should be logged with `helm registry login` command.

```sh
helm registry login xx.xx.xx.xx
```

{{< note >}}
The CA file used by the Harbor must be trusted in the system due to an [issue](https://github.com/helm/helm/issues/6324) in Helm.
{{< /note >}}

### Push Charts to the artifact Repository with the CLI

After logging in, run the `helm chart save` command to save a chart directory that prepares the artifact for pushing.

```sh
helm chart save dummy-chart xx.xx.xx.xx/library/dummy-chart
```

After the chart saves, run the `helm chart push` command to push your charts:

```sh
helm chart push xx.xx.xx.xx/library/dummy-chart:version
```

### Pull Charts from the artifact Repository with the CLI

To pull charts from the the OCI-compatible registry of Harbor, run the `helm chart pull` command just like pulling image using the Docker CLI.

```sh
helm chart pull xx.xx.xx.xx/library/dummy-chart:version
```

### Manage Helm Charts artifacts in Harbor Interface

The charts pushed to the OCI-compatible registry of Harbor are treated like any other type of artifact. You can list, copy, delete, update labels, get details, and add or remove tags for them just like you can for container images.

![chart artifact details](../../../img/chart-artifact-details.png)
