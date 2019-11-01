# Download the Harbor Installer:

You download the installers from the **[official release](https://github.com/goharbor/harbor/releases)** page. Choose either the online or the offline installer. 

- **Online installer:** The online installer downloads the Harbor images from Docker hub. For this reason, the installer is very small in size.

- **Offline installer:** Use the offline installer if the host to which are are deploying Harbor does not have a connection to the Internet. The offline installer contains pre-built images so it is larger than the online installer.

The installation processes are almost the same for both the online and offline installers.

## Download and Unpack the Installer

1. Go to the [Harbor releases page](https://github.com/goharbor/harbor/releases). 
1. Select either the online or offline installer for the version you want to install.
1. Use `tar` to extract the installer package:

   - Online installer:<pre>bash $ tar xvf harbor-online-installer-<em>version</em>.tgz</pre>
   - Offline installer:<pre>bash $ tar xvf harbor-offline-installer-<em>version</em>.tgz</pre>
   
## Next Steps

To prepare your Harbor installation, [Configure the Harbor YML File](configure_yml_file.md).