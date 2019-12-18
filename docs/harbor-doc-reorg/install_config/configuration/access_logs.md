# Access Harbor Logs

By default, registry data is persisted in the host's `/data/` directory.  This data remains unchanged even when Harbor's containers are removed and/or recreated, you can edit the `data_volume` in `harbor.yml` file to change this directory.

In addition, Harbor uses *rsyslog* to collect the logs of each container. By default, these log files are stored in the directory `/var/log/harbor/` on the target host for troubleshooting, also you can change the log directory in `harbor.yml`.