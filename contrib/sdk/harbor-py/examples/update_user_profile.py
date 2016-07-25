#!/usr/bin/env python

import sys
sys.path.append("../")

from harborclient import harborclient

host = "127.0.0.1"
user = "admin"
password = "Harbor12345"

client = harborclient.HarborClient(host, user, password)

# Update user profile
user_id = 2
email = "new@gmail.com"
realname = "new_realname"
comment = "new_comment"
client.update_user_profile(user_id, email, realname, comment)
