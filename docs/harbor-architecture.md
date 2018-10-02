# Harbor Architecture
The following diagram[<sup>1</sup>](#edit-the-architecture-diagram) shows the various services that compose Harbor and how they are related.

![rbac](img/harbor_architecture.png)

## Services
These are the responsibilities of the various services that compose Harbor.

* nginx
    * Reverse proxy for the ui and registry
    * Terminates TLS (if used)
* registry
    * The [Docker Registry](https://docs.docker.com/registry/) is a stateless, highly scalable server side application that stores and lets you distribute Docker images
    * Sends [notifications](https://docs.docker.com/registry/notifications/) to the ui
* ui
    * Serves the user inteface code to browsers
    * Reverse proxy for registry API
    * Sends jobs to job service
    * Retrieves image vulnerability data from clair/postgres
    * Retrieves configuration from adminserver
    * Stores data in mysql
* adminserver
    * Authenticates users for ui and registry
    * Stores configuration
* jobservice
    * Retrieves image vulnerability data from clair/postgres
    * Initiates image vulnerability scans with clair
    * Runs image replication jobs
* mysql
    * Store Harbor data
* clair
    * [Clair](https://coreos.com/clair) is a project for the static analysis of vulnerabilities in appc and docker containers
    * Runs image vulnerability scans
    * Stores data in postgres
    * Sends notifications to the ui
* postgres
    * Stores image vulnerability data

## Edit the Architecture Diagram
To edit this diagram:
1. Go to [draw.io](https://www.draw.io/)
1. Open Existing Diagram and choose this PNG
1. Make your edits
1. When complete choose File > Export as > PNG...
1. Enter Borderwidth: 10
1. Check Transparent Background and Include a copy of my diagram
1. Click Export and overwrite the current PNG
