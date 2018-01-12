#!/usr/bin/python

#used for remove notary signature

import pexpect
import os
import sys
import time
import socket

ip = sys.argv[1]
cmd = "notary -s https://"+ip+":4443 -d ~/.docker/trust/ remove -p "+ip+"/library/tomcat latest"
passwd = "Harbor12345"

child = pexpect.sawn(cmd)
time.sleep(2)

child.expect("username:")
child.sendline("admin")
time.sleep(1)
child.expect("password:")
child.sendline("Harbor12345")
time.sleep(1)
child.expect(":")
child.sendline("Harbor12345")
time.sleep(3)
child.expect(pexpect.EOF)
