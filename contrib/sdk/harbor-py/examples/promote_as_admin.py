#!/usr/bin/env python

import sys
sys.path.append("../")

from harborclient import harborclient

host = "127.0.0.1"
user = "admin"
password = "Harbor12345"

client = harborclient.HarborClient(host, user, password)

# Promote as admin
user_id = 2
client.promote_as_admin(user_id)
