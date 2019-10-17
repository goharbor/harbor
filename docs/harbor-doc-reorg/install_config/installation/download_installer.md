# Download the Harbor Installer:

The binary of the installer can be downloaded from the [release](https://github.com/goharbor/harbor/releases) page. Choose either online or offline installer. 

- **Online installer:** The installer downloads Harbor's images from Docker hub. For this reason, the installer is very small in size.

- **Offline installer:** Use this installer when the host does not have an Internet connection. The installer contains pre-built images so its size is larger.

All installers can be downloaded from the **[official release](https://github.com/goharbor/harbor/releases)** page.

This guide describes the steps to install and configure Harbor by using the online or offline installer. The installation processes are almost the same.


Use *tar* command to extract the package.

Online installer:

```bash
    $ tar xvf harbor-online-installer-<version>.tgz
```

Offline installer:

```bash
    $ tar xvf harbor-offline-installer-<version>.tgz
```