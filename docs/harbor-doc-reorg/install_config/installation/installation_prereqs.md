# Harbor Installation Prerequisites

Harbor is deployed as several Docker containers, and, therefore, can be deployed on any Linux distribution that supports Docker. The target host requires Docker, and Docker Compose to be installed.

### Hardware

|Resource|Capacity|Description|
|---|---|---|
|CPU|minimal 2 CPU|4 CPU is preferred|
|Mem|minimal 4GB|8GB is preferred|
|Disk|minimal 40GB|160GB is preferred|

### Software

|Software|Version|Description|
|---|---|---|
|Docker engine|version 17.06.0-ce+ or higher|For installation instructions, please refer to: [docker engine doc](https://docs.docker.com/engine/installation/)|
|Docker Compose|version 1.18.0 or higher|For installation instructions, please refer to: [docker compose doc](https://docs.docker.com/compose/install/)|
|Openssl|latest is preferred|Generate certificate and keys for Harbor|

### Network ports

|Port|Protocol|Description|
|---|---|---|
|443|HTTPS|Harbor portal and core API will accept requests on this port for https protocol, this port can change in config file|
|4443|HTTPS|Connections to the Docker Content Trust service for Harbor, only needed when Notary is enabled, This port can change in config file|
|80|HTTP|Harbor portal and core API will accept requests on this port for http protocol|