#!/usr/bin/env python

import sys
sys.path.append("../")

from harborclient import harborclient

host = "127.0.0.1"
user = "admin"
password = "Harbor12345"

client = harborclient.HarborClient(host, user, password)

# Get top accessed respositories
print(client.get_top_accessed_repositories())

# Get top accessed respositories with count
count = 1
print(client.get_top_accessed_repositories(count))
