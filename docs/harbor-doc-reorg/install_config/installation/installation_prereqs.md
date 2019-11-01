# Harbor Installation Prerequisites

Harbor is deployed as several Docker containers. You can therefore deploy it on any Linux distribution that supports Docker. The target host requires Docker and Docker Compose to be installed.

## Hardware

The following table lists the minimum and recommended hardware configurations for deploying Harbor.

|Resource|Minimum|Recommended|
|---|---|---|
|CPU|2 CPU|4 CPU|
|Mem|4 GB|8 GB|
|Disk|40 GB|160 GB|

## Software

The following table lists the software versions that must be installed on the target host.

|Software|Version|Description|
|---|---|---|
|Docker Engine|17.06.0-ce+ or higher|For installation instructions, see the  [Docker Engine doc](https://docs.docker.com/engine/installation/)|
|Docker Compose|1.18.0 or higher|For installation instructions, see [Docker Compose doc](https://docs.docker.com/compose/install/)|
|OpenSSL|Latest is preferred|Used to generate the Harbor certificate and keys|

## Network ports

Harbor requires that the following ports be open on the target host.

|Port|Protocol|Description|
|---|---|---|
|443|HTTPS|Harbor interface and core API accept HTTPS requests on this port. You can change this port in the configuration file.|
|4443|HTTPS|Connections to the Docker Content Trust service for Harbor. Only required if Notary is enabled. You can change this port in the configuration file.|
|80|HTTP|Harbor interface and core API accept HTTP requests on this port. You can change this port in the configuration file.|

## What to Do Next ##

To secure the connections the Harbor, you must [Configure HTTPS Access to Harbor](configure_https.md).