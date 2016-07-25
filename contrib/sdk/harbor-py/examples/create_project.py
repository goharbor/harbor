#!/usr/bin/env python

import sys
sys.path.append("../")

from harborclient import harborclient

host = "127.0.0.1"
user = "admin"
password = "Harbor12345"

client = harborclient.HarborClient(host, user, password)

# Create project
project_name = "test-project"
is_public = True
client.create_project(project_name, is_public)
