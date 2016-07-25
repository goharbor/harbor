#!/usr/bin/env python

import sys
sys.path.append("../")

from harborclient import harborclient

host = "127.0.0.1"
user = "admin"
password = "Harbor12345"

client = harborclient.HarborClient(host, user, password)

# Create user
username = "test-username"
email = "test-email@gmail.com"
password = "test-password"
realname = "test-realname"
comment = "test-comment"
client.create_user(username, email, password, realname, comment)
