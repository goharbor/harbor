#!/bin/bash

PRIVILEGED=false

curl -v -X POST $MARATHON_API_URL/v2/apps -H Content-Type:application/json -d \
'{
      "id": "shurenyun-'$TASKENV'-'$SERVICE'",
      "cpus": '$CPUS',
      "mem": '$MEM',
      "instances": '$INSTANCES',
      "constraints": [["hostname", "LIKE", "'$DEPLOYIP'"]],
      "container": {
                     "type": "DOCKER",
                     "docker": {
                                     "image": "'$SERVICE_IMAGE'",
                                     "network": "BRIDGE",
                                     "forcePullImage": '$FORCEPULLIMAGE',
                                     "privileged": '$PRIVILEGED',
                                     "portMappings": [
                                             { "containerPort": 80, "hostPort": 0, "protocol": "tcp"}
                                     ]
                                }
                   },
      "env": {
                    "BAMBOO_PUBLIC": "true",
                    "BAMBOO_PROXY":"true",
                    "BAMBOO_HTTP_PROTOCOL":"true"
             },
      "uris": [
               "'$CONFIGSERVER'/config/demo/config/registry/docker.tar.gz"
       ]
}'

