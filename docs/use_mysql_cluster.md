### Setup
Configuring MySQL Cluster requires working with ``make/db-cluster.cnf`` file in each node.  
In ``db-cluster.cnf``, set up host's ip address and processes that are required MySQL cluster.  

- mgmd: Management node that is sed for configuration and monitoring of the cluster.
- ndbd: Data node. These nodes store the data.
- mysqld: SQL node that connects to all of the data nodes in order to perform data storage and retrieval.

The ``[process name]`` must be unique name.

```
[NODE IP]
mgmd [process name]
ndbd [process name]
mysqld [process name]

[NODE IP]
mgmd [process name]
ndbd [process name]
mysqld [process name]
``` 

The following configuration is two node example.
```
[192.168.56.30]
mgmd mgmd1
ndbd ndbd1
mysqld mysqld1

[192.168.56.31]
ndbd ndbd2
mysqld mysqld2
``` 

### Start up Harbor with MySQL Cluster
In first node, Build, install and bring up Harbor by following command.  
``INITFLAG=true`` must be setting in first node that start up at first.

```
make install GOBUILDIMAGE=golang:1.7.3 COMPILETAG=compile_golangimage CLARITYIMAGE=vmware/harbor-clarity-ui-builder:1.1.2 CLUSTERFLAG=true NODEIP=[NODE IP] INITFLAG=true
```

In second node, bring up Harbor by following command.

```
make install GOBUILDIMAGE=golang:1.7.3 COMPILETAG=compile_golangimage CLARITYIMAGE=vmware/harbor-clarity-ui-builder:1.1.2 CLUSTERFLAG=true NODEIP=[NODE IP] 
```

After all processes are brought up in each node, distribute the data table to MySQL Cluster from first node by following command.

```
make distribute-table
```

