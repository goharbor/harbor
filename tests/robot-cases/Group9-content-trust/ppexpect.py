#!/usr/bin/python

import pexpect
import os
import sys

ip = os.popen("ip a s eth0|grep inet|awk '{print $2}'|awk -F ""/"" '{print $1}'").read().strip('\n')

cmd = "docker push "+ ip +sys.argv[1]
passw = "Harbor12345"
child = pexpect.spawn(cmd)

child.expect(':')
child.sendline(passw)
child.expect(':')
child.sendline(passw)
child.expect(':)')
child.sendline(passw)
child.expect(':)')
child.sendline(passw)
child.expect(pexpect.EOF)
