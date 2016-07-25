#!/usr/bin/env python

import sys
sys.path.append("../")

from harborclient import harborclient

host = "127.0.0.1"
user = "admin"
password = "Harbor12345"

client = harborclient.HarborClient(host, user, password)

# Get project id from name
project_name = "library"
print(client.get_project_id_from_name(project_name))
