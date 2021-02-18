# Chart Migrating Tool

Harbor supports two different ways to storage the chart data.

   1. stored in Harbor registry storage directly via OCI API.
   2. stored in Harbor hosted chartmuseum backend via chartmuseam's API

There is an performance issue in chartmuseam. For example, on my 2 core 8G memory test environment, to get 10000 charts information needs about 10s to return, In 50000 charts situation, It will cause a timeout error.

After version 2.0, Harbor becomes to the OCI registry. So It can storage chart content directly without chartmuseum.

This tool used to migrate the legacy helm charts stored in the chartmuseum backend to Harbor OCI registry backend.

On test environment( 2 core 8G memory), using this tool to migrate 10000 charts needs about 2~3 hours

## Usages

Compile the chart with command

``` sh
docker build -t goharbor/migrate-chart:0.1.0 .
```

Migrate charts run command below:

``` sh
docker run -it --rm -v {{your_chart_data_location}}:/chart_storage -v {{harbor_ca_cert_location}}:/usr/local/share/ca-certificates/harbor_ca.crt  goharbor/migrate-chart:0.1.0 --hostname {{harbor_hostname}} --password {{harbor_admin_password}}
```

* `your_chart_data_location`: The location of your chart storage chart. By default, it's the `chart_storage` dir inside Harbor's `data_volumn`

* `harbor_ca_cert_location`: If Harbor enabled HTTPS, you need to add the `ca_cert` for connecting Harbor

* `harbor_hostname`: The hostname of Harbor

* `harbor_admin_password`: The password of harbor admin user
