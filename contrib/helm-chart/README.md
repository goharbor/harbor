# Harbor Helm Chart

A bootstrap script + Helm Chart to install a fully functioning Harbor registry
on a Kubernetes cluster. This chart is intended to be provider agnostic.
This installation has been tested on AWS (self managed k8s cluster),
Google GKE, and Micorsoft AKS.

## Notes

* assumes the use of blob storage for registry data rather than persistent volume claims
* the design of this installation supports the operation of multiple regional Harbor registries (hence the secrets namespacing) which can be linked via Harbor replication
* in ``values.yaml`` anything with a value of _"placeholder"_ is meant to be overloaded by the bootstrap script which generates keys/secrets/etc and stores them in ``$CLUSTER_SECRETS_DIR`` for secure backup (ex. Keybase, Vault etc)

## Installation

Before you run ``create.sh`` you must bootstrap a namespace called _common_. This namespace is used to store common secrets (such as ssl certificates) that other services in your cluster might need to use (ex. a wildcard ssl certificate *.foo.bar.example.com). You might want to change this so suit your needs (Let's Encrypt etc).

Regarding blob storage, update the _storage_ section of ``values.yaml`` for the
provider you're using (ex. Amazon -> s3, Google -> gcs, Azure -> azure)

Finally once all that has been addressed:
```
# creates the common namespace and injects ssl.crt and ssl.key secret used by Harbor Helm Chart
./common.sh

# examples of running ``create.sh`` against various providers and regions
./create.sh azure ue1
./create.sh aws uw2
./create.sh gcp uw1 /path/to/gcs_keyfile
```


## Credits

This Helm Chart installation was designed by Matt Nuzzaco & DevOps group at [Sight Machine](https://www.sightmachine.com)
