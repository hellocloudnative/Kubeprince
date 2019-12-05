package install

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/pkg/sftp"
	"github.com/wonderivan/logger"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

//版本 （v1.15.6） => 115
func VersionToInt(version string)(int) {
	version = strings.Replace(version,"v","",-1)
	versionArr :=strings.Split(version,".")
	if len(versionArr) >=2 {
		versionArr :=versionArr[0] + versionArr[1]
		if i,err :=strconv.Atoi(versionArr);err ==nil{
			return  i
		}
	}
	return 0
}
//decode output to join token  hash and key
func decodeOutput(output []byte) {
	s0 := string(output)
	slice := strings.Split(s0, "kubeadm join")
	slice1 := strings.Split(slice[1], "Please note")
	//logger.Info("[globals]join command is: %s", slice1[0])
	decodeJoinCmd(slice1[0])
}
//  192.168.0.200:6443 --token 9vr73a.a8uxyaju799qwdjv --discovery-token-ca-cert-hash sha256:7c2e69131a36ae2a042a339b33381c6d0d43887e2de83720eff5359e26aec866 --experimental-control-plane --certificate-key f8902e114ef118304e561c3ecd4d0b543adc226b7a07f675f56564185ffe0c07
func decodeJoinCmd(cmd string) {
	stringSlice := strings.Split(cmd, " ")

	for i, r := range stringSlice {
		switch r {
		case "--token":
			JoinToken = stringSlice[i+1]
		case "--discovery-token-ca-cert-hash":
			TokenCaCertHash = stringSlice[i+1]
		case "--certificate-key":
			CertificateKey = stringSlice[i+1][:64]
		}
	}
}

//提取ip
func IpFormat(host string) string {
	ipAndPort := strings.Split(host, ":")
	return ipAndPort[0]
}

func AddReformat(host string)(string){
	if strings.Index(host,":") == -1{
		host = fmt.Sprintf("%s:22",host)
	}
	return host
}

func ReturnCmd(host, cmd string) string {
	session, _ := Connect(User, Password, PrivateKeyFile, host)
	defer session.Close()
	b, _ := session.CombinedOutput(cmd)
	return string(b)
}

func GetFileSize(url string) int {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[globals] get file size is error： %s", r)
		}
	}()
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	return int(resp.ContentLength)
}

func WatchFileSize(host, filename string, size int) {
	t := time.NewTicker(3 * time.Second) //every 3s check file
	defer t.Stop()
	for {
		select {
		case <-t.C:
			length := ReturnCmd(host, "ls -l "+filename+" | awk '{print $5}'")
			length = strings.Replace(length, "\n", "", -1)
			length = strings.Replace(length, "\r", "", -1)
			lengthByte, _ := strconv.Atoi(length)
			if lengthByte == size {
				t.Stop()
			}
			lengthFloat := float64(lengthByte)
			value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", lengthFloat/oneMBByte), 64)
			logger.Alert("[%s]transfer total size is: %.2f%s", host, value, "MB")
		}
	}
}

//shell命令
func Cmd(host string,cmd string)([]byte){
	logger.Info("[%s]exec cmd is : %s", host, cmd)
	session,err :=Connect(User,Password,PrivateKeyFile,host)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[%s]Error create ssh session failed,%s", host, err)
		}
	}()
	if err !=nil{
		panic(1)
	}
	defer session.Close()
	b,err :=session.CombinedOutput(cmd)
	logger.Debug("[%s]command result is: %s",host,string(b))
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[%s]Error exec command failed: %s", host, err)
			os.Exit(1)
		}
	}()
	if err !=nil{
		panic(1)

	}
	return  b
}

//shell命令
func Cmdout(host string,cmd string)([]byte){
	session,err :=Connect(User,Password,PrivateKeyFile,host)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[%s]Error create ssh session failed,%s", host, err)
		}
	}()
	if err !=nil{
		panic(1)
	}
	defer session.Close()
	b,err :=session.CombinedOutput(cmd)
	logger.Debug("[%s]command result is:[ok]",host)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[%s]Error exec command failed: %s", host, err)
			os.Exit(1)
		}
	}()
	if err !=nil{
		panic(1)

	}
	return  b
}
//判断远程文件是否存在
func RemoteFilExist(host, remoteFilePath string)(bool) {
	remoteFileName :=path.Base(remoteFilePath)
	remoteFileDirName := path.Dir(remoteFilePath)
	remoteFileCommand := fmt.Sprintf("ls -l %s | grep %s | wc -l", remoteFileDirName, remoteFileName)
	data :=bytes.Replace(Cmdout(host,remoteFileCommand),[]byte("\r"), []byte(""), -1)
	data = bytes.Replace(data, []byte("\n"), []byte(""), -1)
	count,err :=strconv.Atoi(string(data))
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[%s]RemoteFilExist:%s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	if count == 0 {
		return false
	} else {
		return true
	}
}
//Copy
func Copy(host, localFilePath, remoteFilePath string) {
	sftpClient, err := SftpConnect(User, Password, PrivateKeyFile, host)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[%s]scpCopy: %s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	defer sftpClient.Close()
	srcFile, err := os.Open(localFilePath)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[%s]scpCopy: %s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remoteFilePath)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[%s]scpCopy: %s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	defer dstFile.Close()
	buf := make([]byte, 100*oneMBByte) //100mb
	totalMB := 0
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		length, _ := dstFile.Write(buf[0:n])
		totalMB += length / oneMBByte
		logger.Alert("[%s]transfer total size is: %d%s", host, totalMB, "MB")
	}
}

func readFile(name string)(string){
	content,err :=ioutil.ReadFile(name)
	if err!=nil{
		logger.Error("[globals] read file err is : %s", err)
		return ""
	}

	return string(content)
}

func sshAuthMethod(password, pkFile string) ssh.AuthMethod {
	var am ssh.AuthMethod
	if password != ""{
		am=ssh.Password(password)
	}else{

		pkData:=readFile(pkFile)
		pk,_ := ssh.ParsePrivateKey([]byte(pkData))
		am = ssh.PublicKeys(pk)
	}
	return am
}
//ssh connect
func Connect(user,password,pkFile,host string) (*ssh.Session, error){
	auth :=[]ssh.AuthMethod{sshAuthMethod(password,pkFile)}
	config :=ssh.Config{
		Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256"},
	}
	clientConfig :=&ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: time.Duration(1) * time.Minute,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) (err error) {
			return nil
		},
	}
	addr := AddReformat(host)
	client,err :=ssh.Dial("tcp",addr,clientConfig)
	if err!=nil{
		return nil,err
	}
	session,err :=client.NewSession()
	if err!=nil{
		return nil,err
	}
	modes :=ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err:=session.RequestPty("xterm", 80, 40, modes);err!=nil{
		return nil,err
	}
	return session,nil
}

//sftp connect
func SftpConnect(user, passwd, pkFile, host string) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, sshAuthMethod(passwd, pkFile))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = AddReformat(host)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return sftpClient, nil
}
//传包
func SendPackage(url string, hosts []string, packName string) {
	pkg := path.Base(url)
	//only http
	isHttp := strings.HasPrefix(url, "http")
	wgetCommand := ""
	if isHttp {
		wgetParam := ""
		if strings.HasPrefix(url, "https") {
			wgetParam = "--no-check-certificate"
		}
		wgetCommand = fmt.Sprintf(" wget %s ", wgetParam)
	}
	//wget下载
	remoteCmd := fmt.Sprintf("cd /root &&  %s %s && tar zxf %s", wgetCommand, url, pkg)
	//本地包
	localCmd := fmt.Sprintf("cd /root && rm -rf %s && tar zxf %s ", packName, pkg)
	kubeLocal := fmt.Sprintf("/root/%s", pkg)
	var kubeCmd string
	if packName == "kube" {
		kubeCmd = "cd /root/kube/&& sh init.sh"
	} else {
		kubeCmd = fmt.Sprintf("cd /root/%s && docker load -i images.tar", packName)
	}
	var wm sync.WaitGroup
	for _, host := range hosts {
		wm.Add(1)
		go func(host string) {
			defer wm.Done()
			logger.Debug("[%s]please wait for decompressing ......", host)
			if RemoteFilExist(host, kubeLocal) {
				logger.Warn("[%s]SendPackage: file is exist", host)
				Cmdout(host, localCmd)
			} else {
				if isHttp {
					go WatchFileSize(host, kubeLocal, GetFileSize(url))
					Cmdout(host, remoteCmd)
				} else {
					Copy(host, url, kubeLocal)
					Cmdout(host, localCmd)
				}
			}
			Cmdout(host, kubeCmd)
		}(host)
	}
	wm.Wait()
}

func KubeadmConfigInstall(){
	var templateData string
	if  KubeadmFile == ""{
		templateData =string(Template())

	}else {
		fileData, err := ioutil.ReadFile(KubeadmFile)
		defer func() {
			if r := recover(); r != nil {
				logger.Error("[globals]template file read failed:", err)
			}
		}()
		if err != nil {
			panic(1)
		}
		templateData = string(TemplateFromTemplateContent(string(fileData)))
	}
	cmd := "echo \"" + templateData + "\" > /root/kube/conf/kubeadm-config.yaml"
	Cmdout(Masters[0], cmd)
}