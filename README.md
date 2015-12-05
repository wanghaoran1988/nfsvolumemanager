# nfsvolumemanager
**This is tool is designed for create nfs pv**

**Install**
* clone the code loally
* godep go build
* the output binary file :nfsvolumemanager

**Run on the nfs server**

* Start nfs server :  
   yum -y install nfs-utils nfs-utils-lib  
   systemctl start rpcbind  
   systemctl start nfs  
* start nfsvolumemanager  
   sudo nohup ./nfsvolumemanager >nfsvolumemage.log 2>&1 &  

**Create a pv**  
ex: create a pv "haowangpv"
   oc create -f http://<ip>:8000/volumes/haowangpv  
   
**Notes**  
    when invoke http://<nfs_server>:8000/volumes/{volumename},This will create a directory <volumename> under /nfs/volumes/ in the nfs server, and the directory permision will be 777, the owner will be nfsnobody:nfsnobody, after created the folder , will update the nfs configuration file /etc/exports ,then invoke "exportfs -r" command. this API will return a PV definition json file that can be used to create a pv.  
