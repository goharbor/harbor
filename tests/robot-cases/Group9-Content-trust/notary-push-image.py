#!/usr/bin/python

import pexpect
import os
import sys
import time
import socket

ip = socket.gethostbyname(socket.gethostname())
cmd = "docker push "+ip+"/library/tomcat:latest"
passw = "Harbor12345"
child = pexpect.spawn(cmd)
time.sleep(5)

child.expect(':')
child.sendline(passw)
time.sleep(1)
child.expect(':')
child.sendline(passw)
time.sleep(1)
child.expect(':')
child.sendline(passw)
time.sleep(1)
child.expect(':')
child.sendline(passw)
time.sleep(1)
child.expect(pexpect.EOF)
