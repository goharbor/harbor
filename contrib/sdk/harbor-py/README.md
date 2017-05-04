## Introduction

**harbor-py** is the native python SDK and cli tool for Harbor. In the **harborclient** directory, ```harborsdk.py``` is the python SDK for Harbor, ```harborcli.py``` is the command line tool for Harbor. Also, you can develop your own cli tool based on the SDK. The SDK supports all APIs of Harbor. 

## Installation

Run command below:
```
sudo python setup.py install 
```

## Usage

First, export Harbor environment variables under your bash console.

```
export HARBOR_HOSTNAME=HarborIP
export HARBOR_USER=username
export HARBOR_PASSWORD=password
```

If your Harbor uses https, export one more enviroment variable:

```
export HARBOR_URL_PROTOCOL=https
```

```harborcli``` has a lot of subcommands, to get help infomation:

```
harborcli <subcommand> -h
```

Some examples are listed below.

**Getting project information:**
```
harborcli get-projects -n <project_name> -i <public> -o <owner> -p <page> -s <page_size>
```

**Creating a new user:**
```
harborcli create-user <username> <email> <password> <realname> -c <comment>
```

**Creating a new project:**
```
harborcli create-project <project_name> <public> -e <enable_content_trust> \
                                             -p <prevent_vulnerable_images_from_running> \
                                             -s <prevent_vulnerable_images_from_running_severity> \
                                             -a <automatically_scan_images_on_push>
```

For more usage, please refer to **harborcli -h**.

## Contribution

If you have any suggestions, feel free to submit [issues](https://github.com/vmware/harbor/issues) or send us pull requests.
