# Harbor Backup and Restore Guide

When the container registry storage is local file system, the straightforward way to backup a standalone Harbor data is to backup the /data directory and recover it to /data directory.
For some special cases, if you want more efficient way to backup the Harbor data, you can refer the guide following. 

## Backup Harbor data

1. Login to the Harbor host, find a disk which has enough space, create a backup directory in it, for exmaple, /backup
2. Download script from https://github.com/goharbor/harbor/blob/backup_restore/tools/harbor-backup.sh and copy it to /backup
3. Shutdown the Harbor instance  
```
docker-compose down -v
```
4. Backup data
```
cd /backup
ls
harbor-backup.sh
```

Stop all running containers. Backup all data

```
./harbor-backup.sh 
```
Or only backup database data when Harbor storage is NFS, GCP or S3
```
./harbor-backup.sh --dbonly
```
4. After the backup complete, there is a harbor.tgz file in /backup, it is the backup data, copy it to the backup storage.
```
ls /backup
harbor-backup.sh    harbor.tgz
```
5. Start Harbor 
```
docker-compose up -d
```
## Restore Harbor data
Install Harbor with the same version and same install options. Login to the Harbor host.
Copy the backup data file harbor.tgz to the directory /restore
Download script from https://github.com/goharbor/harbor/blob/backup_restore/tools/harbor-restore.sh and copy it to /restore

1. Shutdown the Harbor instance
```
docker-compose down -v
```
2. Restore data
```
cd /restore
ls
harbor-restore.sh   harbor.tgz
```
Stop all running containers. Restore all data
```
./harbor-restore.sh 
```
Or only restore database data when the Harbor registry storage is NFS, GCS, S3 or Azure.
```
./harbor-restore.sh --dbonly
```
3. Start Harbor
```
docker-compose up -d
```
4. Remove all unused data in /restore
```
rm -rf /restore
```