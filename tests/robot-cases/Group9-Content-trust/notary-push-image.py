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

try:
    child.expect(r'Enter passphrase for new root key with ID.*:')
    child.sendline(passw)
    time.sleep(1)
    child.expect(r'Repeat passphrase for new root key with ID.*:')
    child.sendline(passw)
    time.sleep(1)
    child.expect(r'Enter passphrase for new repository key with ID.*:')
    child.sendline(passw)
    time.sleep(1)
    child.expect(r'Repeat passphrase for new repository key with ID.*:')
    child.sendline(passw)
    time.sleep(1)
    child.expect(pexpect.EOF)
finally:
    print child.before
