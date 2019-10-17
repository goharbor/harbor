# Run the Installer Script

Once **harbor.yml** and storage backend (optional) are configured, install and start Harbor using the `install.sh` script.  Note that it may take some time for the online installer to download Harbor images from Docker hub.

##### Default installation (without Notary/Clair)

Harbor has integrated with Notary and Clair (for vulnerability scanning). However, the default installation does not include Notary or Clair service.

``` sh
    $ sudo ./install.sh
```

If everything worked properly, you should be able to open a browser to visit the admin portal at `http://reg.yourdomain.com` (change `reg.yourdomain.com` to the hostname configured in your `harbor.yml`). Note that the default administrator username/password are admin/Harbor12345.

Log in to the admin portal and create a new project, e.g. `myproject`. You can then use docker commands to login and push images (By default, the registry server listens on port 80):

```sh
$ docker login reg.yourdomain.com
$ docker push reg.yourdomain.com/myproject/myrepo:mytag
```

**IMPORTANT:** The default installation of Harbor uses _HTTP_ - as such, you will need to add the option `--insecure-registry` to your client's Docker daemon and restart the Docker service.

##### Installation with Notary
To install Harbor with Notary service, add a parameter when you run `install.sh`:

```sh
    $ sudo ./install.sh --with-notary
```

**Note**: For installation with Notary the parameter **ui_url_protocol** must be set to "https". For configuring HTTPS please refer to the following sections.

More information about Notary and Docker Content Trust, please refer to [Docker's documentation](https://docs.docker.com/engine/security/trust/content_trust/).

##### Installation with Clair

To install Harbor with Clair service, add a parameter when you run `install.sh`:

```sh
    $ sudo ./install.sh --with-clair
```

For more information about Clair, please refer to Clair's documentation:
`https://coreos.com/clair/docs/2.0.1/`

##### Installation with chart repository service

To install Harbor with chart repository service, add a parameter when you run ```install.sh```:

```sh
    $ sudo ./install.sh --with-chartmuseum
```

**Note**: If you want to install Notary, Clair and chart repository service, you must specify all the parameters in the same command:

```sh
    $ sudo ./install.sh --with-notary --with-clair --with-chartmuseum
```

For information on how to use Harbor, please refer to **[User Guide of Harbor](user_guide.md)** .

