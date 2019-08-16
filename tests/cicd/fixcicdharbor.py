#!/usr/bin/env python3

import yaml,os
print("fix cicd harbor")

config=yaml.safe_load(open('/input/harbor.yml'))
config['hostname']=os.environ.get('IP', '127.0.0.1')
config['data_volume']=os.environ.get('data_volume', '/data')
config['http']['port']=os.environ.get('HTTP_PORT', 80)
config['https']={}
config['https']['port']=os.environ.get('HTTPS_PORT', 443)
config['https']['certificate']=os.environ.get('certificate', '/cert/server.crt')
config['https']['private_key']=os.environ.get('private_key', '/cert/server.key')
config['log']['local']['location']=os.environ.get('data_volume', '/data')+'/logs'

yaml.dump(config, open('/input/harbor.yml', 'w+'))

versions=yaml.safe_load(open('versions'))
versions['VERSION_TAG']=os.environ.get('TAG', 'dev')
yaml.dump(versions, open('versions', 'w+'))

import main
try:
    main.main()
except SystemExit as e:
    if e.code != 0:
        raise e

compose=yaml.safe_load(open('/compose_location/docker-compose.yml'))
NAMESPACE=os.environ.get('NAMESPACE', 'goharbor')
for s in compose['services'].values():
    s['image']=s['image'].replace('goharbor'+"/", NAMESPACE+'/')
    s['container_name']=s['container_name']+"-"+versions['VERSION_TAG']
    if isinstance(s['networks'], dict):
        nn={}
        for n in s['networks']:
            nn[n+"-"+versions['VERSION_TAG']]=s['networks'][n]
        s['networks']=nn
    else:
        nn=[]
        for n in s['networks']:
            nn.append(n+"-"+versions['VERSION_TAG'])
        s['networks']=nn
nn={}
for n in compose['networks']:
    nn[n+"-"+versions['VERSION_TAG']]=compose['networks'][n]
compose['networks']=nn
yaml.dump(compose, open('/compose_location/docker-compose.yml', 'w+'))
