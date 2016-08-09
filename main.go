package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	. "github.com/emicklei/go-restful"
	"github.com/golang/glog"
)

// This example shows how to a Route that matches the "tail" of a path.
// Requires the use of a CurlyRouter and the star "*" path parameter pattern.
//
// GET http://localhost:8080/basepath/some/other/location/test.xml
const (
	baseVolumePath             = "/nfs/volumes/"
	fileMode       os.FileMode = 0777
	nfsconfigfile              = "/etc/exports"
)

type PV struct {
	APIVersion string    `json:"apiVersion"`
	Kind       string    `json:"kind"`
	Meta       *MetaData `json:"metadata"`
	Spec       *Spec     `json:"spec"`
}
type MetaData struct {
	Name string `json:"name"`
}
type Spec struct {
	Capacity    *Capacity `json:"capacity"`
	AccessModes []string  `json:"accessModes"`
	Nfs         *NFS      `json:"nfs"`
	Policy      string    `json:"persistentVolumeReclaimPolicy"`
}
type Capacity struct {
	Storage string `json:"storage"`
}
type NFS struct {
	Path   string `json:"path"`
	Server string `json:"server"`
}

func main() {
	//should be run as root
	if user := os.Getuid(); user != 0 {
		println("Should run as root user")
		os.Exit(1)
	}
	//check where the baseVloumepath exits or not
	// make sure the baseVolumePath exists
	info, err := os.Stat(baseVolumePath)
	if err != nil {
		println(baseVolumePath + " not exits, will create ")
		os.MkdirAll(baseVolumePath, fileMode)
		os.Chown(baseVolumePath, 65534, 65534)
		os.Chmod(baseVolumePath, fileMode)
	} else if !info.IsDir() {
		println(baseVolumePath + " exits, but is a file ,need a directory")
		os.Exit(1)
	}
	DefaultContainer.Router(CurlyRouter{})
	ws := new(WebService)
	ws.Route(ws.GET("/").To(helloHandler))
	ws.Route(ws.GET("/volumes/{name:*}").To(volumeHandler))
	Add(ws)
	println("nfs volume manager start at  http://localhost:8000/")
	http.ListenAndServe(":8000", nil)
}
func helloHandler(req *Request, resp *Response) {

	io.WriteString(resp, "This is used to create nfs volume on server, please use /volumes/{name}/ to create a volume for you, volumes created will be under: "+baseVolumePath)
}
func volumeHandler(req *Request, resp *Response) {
	volumeName := req.PathParameter("name")
	volumePath := baseVolumePath + volumeName
	//check whether the directory exits or not, if exits , will delete all the file under this directory
	_, err := os.Stat(volumePath)
	if err != nil {
		println("Volume :" + volumeName + " not exits , will create one")
	} else {
		println(volumePath + " exists , will delete it first")
		os.RemoveAll(volumePath)
	}
	os.MkdirAll(volumePath, fileMode)
	os.Chown(volumePath, 65534, 65534)
	os.Chmod(volumePath, fileMode)
	updateNFSConfig(volumePath)
	freshNFSConfig()
	server := strings.Split(req.Request.Host, ":")[0]
	//var storage string
	//req.Request.ParseForm()
	//if storage := req.Request.Form.Get("storage"); len(storage) == 0 {
	//	storage = "512Mi"
	//}
	storage := "2Gi"
	//var policy string
	//if policy := req.HeaderParameter("policy"); len(policy) == 0 {
	//	println("policy:" + policy)
	//	policy = "Recycle"
	//}
	policy := "Recycle"
	pvstring := getPVString(volumeName, server, volumePath, storage, policy)
	//io.WriteString(resp, "Volume created: "+volumePath)
	println(pvstring)
	io.WriteString(resp, pvstring)
}
func getPVString(name, server, hostPath, storage, policy string) string {
	println("storage: " + storage)
	pv := &PV{}
	pv.APIVersion = "v1"
	pv.Kind = "PersistentVolume"
	meta := &MetaData{
		Name: name,
	}
	pv.Meta = meta
	capacity := &Capacity{
		Storage: storage,
	}
	accessModes := []string{"ReadWriteOnce"}
	nfs := &NFS{
		Path:   hostPath,
		Server: server,
	}
	spec := &Spec{
		Capacity:    capacity,
		AccessModes: accessModes,
		Nfs:         nfs,
		Policy:      policy,
	}
	pv.Spec = spec
	if b, err := json.Marshal(pv); err == nil {
		fmt.Println("================struct 到json str==")
		fmt.Println(string(b))
		return string(b)
	} else {
		fmt.Println("failed to marshal the pv :" + name)
		return ""
	}
}
func updateNFSConfig(path string) {
	if VolumeExists(path) {
		println("volume : " + path + "  exits in file :" + nfsconfigfile)
	} else {
		println("volume : "+path+"  not exits in file :"+nfsconfigfile, "will append to it.")
		f, err := os.OpenFile(nfsconfigfile, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		line := path + " *(rw)\n"
		f.WriteString(line)
	}
}
func VolumeExists(path string) bool {
	//input, err := os.Open(nfsconfigfile) // For read access.
	input, err := ioutil.ReadFile(nfsconfigfile)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		if strings.Contains(line, path) {
			println(path + " exits in " + nfsconfigfile)
			return true
		}
	}
	return false
}
func freshNFSConfig() {
	cmd := exec.Command("exportfs", "-r")
	err := cmd.Run()
	if err != nil {
		glog.V(4).Infof("Error from local command execution: %v", err)
	}
}
