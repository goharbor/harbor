# Expand the Harddisk of Virtual Appliance

If you install Harbor with OVA, you can refer to this guide  to expand the size of harddisk which Harbor's persistent data is stored in.  

1. Add New Harddisk to VM  

	(1) Log in vSphere web client. Power off Harbor's virtual appliance.  
	(2) Right click on the VM and select "Edit Settings".  
	(3) Select "New Hard Disk", and click "OK".  
	(4) Power on the VM.  

2. Expand Harddisk using LVM  

	Login from the console of the virtual appliance and run the following commands:  

	(1) Check the current size of "/data":  
	```sh
	df -h /data
	```
	
	![lvm](img/lvm/size_of_data_01.png)
	
	(2) Find the new harddisk, e.g. "/dev/sdc". Replace all "/dev/sdc" with your disk in the following commands.  
	```sh
	fdisk -l
	```
	
	![lvm](img/lvm/find_the_new_harddisk.png)
	
	(3) Create new physical volume:  
	```sh
	pvcreate /dev/sdc
	```
	
	(4) Check the volume group:  
	```sh
	vgdisplay
	```
	
	![lvm](img/lvm/vg_01.png)
	
	(5) Expand the volume group:
	```sh
	vgextend data1_vg /dev/sdc
	```
	
	(6) Check the volume group again, note the number of "Free PE":  
	```sh
	vgdisplay
	```
	
	![lvm](img/lvm/vg_02.png)
	
	(7) Check the logical volume:
	```sh
	lvdisplay
	```
	
	![lvm](img/lvm/lv_01.png)
	
	(8) Resize the logical volume, replace "n" with your "Free PE" in step (6):  
	```sh
	lvresize -l +n /dev/data1_vg/data
	```
	
	![lvm](img/lvm/resize_lv.png)
	
	(9) Check the logical volume again, note the change of "LV Size":
	```sh
	lvdisplay
	```
	
	![lvm](img/lvm/lv_02.png)
	
	(10) Resize the file system:
	```sh
	resize2fs /dev/data1_vg/data
	```
	
	(11) Check the size "/data" again:
	```sh
	df -h /data
	```
	
	![lvm](img/lvm/size_of_data_02.png)

After that, your disk should be expanded successfully.