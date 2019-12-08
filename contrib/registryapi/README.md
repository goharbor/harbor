# registryapi
api for docker registry by token authorization

+ a simple api class which lies in registryapi.py, which simulates the interactions 
between docker registry and the vendor authorization platform like harbor.
```
usage:
from registry import RegistryApi
api = RegistryApi('username', 'password', 'http://www.your_registry_url.com/')
repos = api.getRepositoryList()
tags = api.getTagList('public/ubuntu')
manifest = api.getManifest('public/ubuntu', 'latest')
res = api.deleteManifest('public/ubuntu', '23424545**4343')

```

+ a simple client tool based on api class, which contains basic read and delete 
operations for repo, tag, manifest
```
usage:
./cli.py --username username --password passwrod --registry_endpoint http://www.your_registry_url.com/ target action params

target can be: repo, tag, manifest
action can be: list, get, delete
params can be: --repo --ref --tag

more see: ./cli.py -h

```
