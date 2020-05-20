#!/usr/bin/python

import argparse
import json
import requests
import sys

parser = argparse.ArgumentParser()
parser.add_argument("-H", "--host", help="The Harbor server need to config")
parser.add_argument("-u", "--user", default="admin", help="The Harbor username")
parser.add_argument("-p", "--password", default="Harbor12345", help="The Harbor password")
parser.add_argument("-c", "--config", nargs='+', help="The configure settings <key>=<value>, it can take more than one configures")
args = parser.parse_args()
reqJson = {}
for item in args.config :
    configs = item.split("=", 1)
    key = configs[0].strip()
    value = configs[1].strip()
    if value.lower() in ['true', 'yes', '1'] :
        reqJson[key] = True
    elif value.lower() in ['false', 'no', '0'] :
        reqJson[key] = False
    elif value.isdigit() :
        reqJson[key] = int(value)
    else:
        reqJson[key] = value

# Sample Basic Auth Url with login values as username and password
url = "https://"+args.host+"/api/v2.0/configurations"
user = args.user
passwd = args.password

# Make a request to the endpoint using the correct auth values
auth_values = (user, passwd)
session = requests.Session()
session.verify = False
data = json.dumps(reqJson)
headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
response = session.put(url, auth=auth_values, data=data, headers=headers)

# Convert JSON to dict and print
if response.status_code == 200 :
    print("Configure setting success")
    print("values:"+data)
    sys.exit(0)
else:
    print("Failed with http return code:"+ str(response.status_code))
    sys.exit(1)
