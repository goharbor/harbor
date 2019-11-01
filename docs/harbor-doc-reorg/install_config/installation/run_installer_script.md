# Run the Installer Script

Once you have configured **harbor.yml** optionally set up a storage backend, you install and start Harbor by using the `install.sh` script. Note that it might take some time for the online installer to download all of the `Harbor images from Docker hub.

You can install Harbor in different configurations:

- Just Harbor, without Notary, Clair, or Chart Repository Service
- Harbor with Notary
- Harbor with Clair
- Harbor with Chart Repository Service
- Harbor with two or all three of Notary, Clair, and Chart Repository Service

## Default installation without Notary, Clair, or Chart Repository Service

The default Harbor installation does not include Notary or Clair service.

``` sh
    $ sudo ./install.sh
```

If the installation succeeds, you can open a browser to visit the Harbor Portal at `http://reg.yourdomain.com`, changing `reg.yourdomain.com` to the hostname that you configured in `harbor.yml`. If you did not change them, the default administrator username and password are `admin` and `Harbor12345`.

Log in to the admin portal and create a new project, for example, `myproject`. You can then use docker commands to log in and push images to Harbor. By default, the registry server listens on port 80:

```sh
$ docker login reg.yourdomain.com
$ docker push reg.yourdomain.com/myproject/myrepo:mytag
```

**IMPORTANT:** If your installation of Harbor uses HTTP, you must add the option `--insecure-registry` to your client's Docker daemon and restart the Docker service.

## Installation with Notary

To install Harbor with the Notary service, add the `--with-notary` parameter when you run `install.sh`:

```sh
    $ sudo ./install.sh --with-notary
```

**Note**: For installation with Notary, you must use Harbor with HTTPS.

For more information about Notary and Docker Content Trust, see [Content Trust](https://docs.docker.com/engine/security/trust/content_trust/) in the Docker documentation.

## Installation with Clair

To install Harbor with Clair service, add the `--with-clair` parameter when you run `install.sh`:

```sh
    $ sudo ./install.sh --with-clair
```

For more information about Clair, see the [Clair documentation](https://coreos.com/clair/docs/2.0.1/).

## Installation with Chart Repository Service 

To install Harbor with chart repository service, add the `--with-chartmuseum` parameter when you run ```install.sh```:

```sh
    $ sudo ./install.sh --with-chartmuseum
```

## Installation with Notary, Clair, and Chart Repository Service

If you want to install all three of Notary, Clair and chart repository service, you must specify all of the parameters in the same command:

```sh
    $ sudo ./install.sh --with-notary --with-clair --with-chartmuseum
```

## What to Do Next ##

If installation succeeds, continue to set up Harbor by following the instructions in [Post-Installation Configuration](install_config/configuration/_index.md) and [Initial Configuration in the Harbor UI](install_config/configuration/initial_config_ui.md).

If installation fails, see [Troubleshooting Harbor Installation
](install_config/installation/troubleshoot_installation.md).
