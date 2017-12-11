import base64
import paramiko
key = paramiko.RSAKey(data=base64.b64decode(b'AAA...'))
client = paramiko.SSHClient()
client.get_host_keys().add('10.162.29.254', 'ssh-rsa', key)
client.connect('10.162.29.254', username='ubuntu', password='vmware')
stdin, stdout, stderr = client.exec_command('ls')
for line in stdout:
    print('... ' + line.strip('\n'))
client.close()