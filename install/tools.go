package install

import (
	"crypto/tls"
	"fmt"
	"github.com/wonderivan/logger"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"time"
)


func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// 返回/etc/hosts记录
func getApiserverHost(ipAddr string) (host string) {
	return fmt.Sprintf("%s %s", ipAddr, ApiServer)
}

//VersionToInt v1.15.6  => 115
func VersionToInt(version string) int {
	// v1.15.6  => 1.15.6
	version = strings.Replace(version, "v", "", -1)
	versionArr := strings.Split(version, ".")
	if len(versionArr) >= 2 {
		versionStr := versionArr[0] + versionArr[1]
		if i, err := strconv.Atoi(versionStr); err == nil {
			return i
		}
	}
	return 0
}

func GetRemoteHostName(hostIP string) string {
	hostName := SSHConfig.CmdToString(hostIP, "hostname", "")
	return strings.ToLower(hostName)
}

//根据yaml转换kubeadm结构
func KubeadmDataFromYaml(context string) *KubeadmType {
	yamls := strings.Split(context, "---")
	if len(yamls) > 0 {
		for _, y := range yamls {
			cfg := strings.TrimSpace(y)
			if cfg == "" {
				continue
			} else {
				kubeadm := &KubeadmType{}
				if err := yaml.Unmarshal([]byte(cfg), kubeadm); err == nil {
					//
					if kubeadm.Kind == "ClusterConfiguration" {
						if kubeadm.Networking.DnsDomain == "" {
							kubeadm.Networking.DnsDomain = "cluster.local"
						}
						return kubeadm
					}
				}
			}
		}
	}
	return nil
}

type KubeadmType struct {
	Kind      string `yaml:"kind,omitempty"`
	ApiServer struct {
		CertSANs []string `yaml:"certSANs,omitempty"`
	} `yaml:"apiServer"`
	Networking struct {
		DnsDomain string `yaml:"dnsDomain,omitempty"`
	} `yaml:"networking"`
}

//获取kubeprince绝对路径
func FetchubePrinceAbsPath() string {
	ex, _ := os.Executable()
	exPath, _ := filepath.Abs(ex)
	return exPath
}

//VersionToIntAll v1.19.1 ==> 1191
func VersionToIntAll(version string) int {
	version = strings.Replace(version, "v", "", -1)
	arr := strings.Split(version, ".")
	if len(arr) >= 3 {
		str := arr[0] + arr[1] + arr[2]
		if i, err := strconv.Atoi(str); err == nil {
			return i
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

//func ReturnCmd(host, cmd string) string {
//	session, _ := Connect(User, Password, PrivateKeyFile, host)
//	defer session.Close()
//	b, _ := session.CombinedOutput(cmd)
//	return string(b)
//}

//func GetFileSize(url string) int {
//	tr := &http.Transport{
//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
//	}
//
//	client := &http.Client{Transport: tr}
//	resp, err := client.Get(url)
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[globals] get file size is error： %s", r)
//		}
//	}()
//	if err != nil {
//		panic(err)
//	}
//	resp.Body.Close()
//	return int(resp.ContentLength)
//}
//
//func WatchFileSize(host, filename string, size int) {
//	t := time.NewTicker(3 * time.Second) //every 3s check file
//	defer t.Stop()
//	for {
//		select {
//		case <-t.C:
//			length := ReturnCmd(host, "ls -l "+filename+" | awk '{print $5}'")
//			length = strings.Replace(length, "\n", "", -1)
//			length = strings.Replace(length, "\r", "", -1)
//			lengthByte, _ := strconv.Atoi(length)
//			if lengthByte == size {
//				t.Stop()
//			}
//			lengthFloat := float64(lengthByte)
//			value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", lengthFloat/oneMBByte), 64)
//			logger.Alert("[%s]transfer total size is: %.2f%s", host, value, "MB")
//		}
//	}
//}

//shell命令
//func Cmd(host string,cmd string)([]byte){
//	session,err :=Connect(User,Password,PrivateKeyFile,host)
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[%s]Error create ssh session failed,%s", host, err)
//		}
//	}()
//	if err !=nil{
//		panic(1)
//	}
//	defer session.Close()
//	b,err :=session.CombinedOutput(cmd)
//	logger.Debug("[%s]command result is: %s",host,string(b))
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[%s]Error exec command failed: %s", host, err)
//			os.Exit(1)
//		}
//	}()
//	if err !=nil{
//		panic(1)
//
//	}
//	return  b
//}

//shell命令
//func Cmdout(host string,cmd string)([]byte){
//	session,err :=Connect(User,Password,PrivateKeyFile,host)
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[%s]Error create ssh session failed,%s", host, err)
//		}
//	}()
//	if err !=nil{
//		panic(1)
//	}
//	defer session.Close()
//	b,err :=session.CombinedOutput(cmd)
//	logger.Debug("[%s]command result is:[ok]",host)
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[%s]Error exec command failed: %s", host, err)
//			os.Exit(1)
//		}
//	}()
//	if err !=nil{
//		panic(1)
//
//	}
//	return  b
//}

//判断远程文件是否存在
//func RemoteFilExist(host, remoteFilePath string)(bool) {
//	remoteFileName :=path.Base(remoteFilePath)
//	remoteFileDirName := path.Dir(remoteFilePath)
//	remoteFileCommand := fmt.Sprintf("ls -l %s | grep %s | wc -l", remoteFileDirName, remoteFileName)
//	data :=bytes.Replace(Cmdout(host,remoteFileCommand),[]byte("\r"), []byte(""), -1)
//	data = bytes.Replace(data, []byte("\n"), []byte(""), -1)
//	count,err :=strconv.Atoi(string(data))
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[%s]RemoteFilExist:%s", host, err)
//		}
//	}()
//	if err != nil {
//		panic(1)
//	}
//	if count == 0 {
//		return false
//	} else {
//		return true
//	}
//}

//Copy
//func Copy(host, localFilePath, remoteFilePath string) {
//	sftpClient, err := SftpConnect(User, Password, PrivateKeyFile, host)
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[%s]scpCopy: %s", host, err)
//		}
//	}()
//	if err != nil {
//		panic(1)
//	}
//	defer sftpClient.Close()
//	srcFile, err := os.Open(localFilePath)
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[%s]scpCopy: %s", host, err)
//		}
//	}()
//	if err != nil {
//		panic(1)
//	}
//	defer srcFile.Close()
//
//	dstFile, err := sftpClient.Create(remoteFilePath)
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[%s]scpCopy: %s", host, err)
//		}
//	}()
//	if err != nil {
//		panic(1)
//	}
//	defer dstFile.Close()
//	buf := make([]byte, 100*oneMBByte) //100mb
//	totalMB := 0
//	for {
//		n, _ := srcFile.Read(buf)
//		if n == 0 {
//			break
//		}
//		length, _ := dstFile.Write(buf[0:n])
//		totalMB += length / oneMBByte
//		logger.Alert("[%s]transfer total size is: %d%s", host, totalMB, "MB")
//	}
//}

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

var message string

func pkgUrlCheck(pkgUrl string) bool {
	if !strings.HasPrefix(pkgUrl, "http") && !FileExist(pkgUrl) {
		message = ErrorFileNotExist
		logger.Error(message + "please check where your PkgUrl is right?")
		return true
	}
	// 判断PkgUrl, 有http前缀时, 下载的文件如果小于400M ,则报错.
	return strings.HasPrefix(pkgUrl, "http") && !downloadFileCheck(pkgUrl)
}

func downloadFileCheck(pkgUrl string) bool {
	u, err := url.Parse(pkgUrl)
	if err != nil {
		return false
	}
	if u != nil {
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			logger.Error(ErrorPkgUrlNotExist, "please check where your PkgUrl is right?")
			return false
		}
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		_, err = client.Do(req)
		if err != nil {
			logger.Error(err)
			return false
		}
		/*
			if tp := resp.Header.Get("Content-Type"); tp != "application/x-gzip" {
				logger.Error("your pkg url is  a ", tp, "file, please check your PkgUrl is right?")
				return false
			}
		*/

		//if resp.ContentLength < MinDownloadFileSize { //判断大小 这里可以设置成比如 400MB 随便设置一个大小
		//	logger.Error("your pkgUrl download file size is : ", resp.ContentLength/1024/1024, "m, please check your PkgUrl is right")
		//	return false
		//}
	}
	return true
}



// ExitOSCase is
func ExitInitCase() bool {
	// 重大错误直接退出, 不保存配置文件
	if len(Masters) == 0 {
		message = ErrorMasterEmpty
	}
	if Version == "" {
		message += ErrorVersionEmpty
	}
	// 用户不写 --passwd, 默认走pk, 秘钥如果没有配置ssh互信, 则验证ssh的时候报错. 应该属于preRun里面
	// first to auth password, second auth pk.
	// 如果初始状态都没写, 默认都为空. 报这个错
	//if SSHConfig.Password == "" && SSHConfig.PkFile == "" {
	//	message += ErrorMessageSSHConfigEmpty
	//}
	if message != "" {
		logger.Error(message + "please check your command is ok?")
		return true
	}

	return pkgUrlCheck(PkgUrl)
}

//sftp connect
//func SftpConnect(user, passwd, pkFile, host string) (*sftp.Client, error) {
//	var (
//		auth         []ssh.AuthMethod
//		addr         string
//		clientConfig *ssh.ClientConfig
//		sshClient    *ssh.Client
//		sftpClient   *sftp.Client
//		err          error
//	)
//	// get auth method
//	auth = make([]ssh.AuthMethod, 0)
//	auth = append(auth, sshAuthMethod(passwd, pkFile))
//
//	clientConfig = &ssh.ClientConfig{
//		User:    user,
//		Auth:    auth,
//		Timeout: 30 * time.Second,
//		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
//			return nil
//		},
//	}
//
//	// connet to ssh
//	addr = AddReformat(host)
//
//	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
//		return nil, err
//	}
//
//	// create sftp client
//	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
//		return nil, err
//	}
//
//	return sftpClient, nil
//}

//func SendPackage() {
//	all := append(p.Masters, p.Nodes...)
//	pkg := path.Base(p.NewPkgUrl)
//	//only http
//	isHttp := strings.HasPrefix(url, "http")
//	wgetCommand := ""
//	if isHttp {
//		wgetParam := ""
//		if strings.HasPrefix(url, "https") {
//			wgetParam = "--no-check-certificate"
//		}
//		wgetCommand = fmt.Sprintf(" wget %s ", wgetParam)
//	}
//	//wget下载
//	remoteCmd := fmt.Sprintf("cd /root &&  %s %s && tar zxf %s", wgetCommand, url, pkg)
//	//本地包
//	localCmd := fmt.Sprintf("cd /root && rm -rf %s && tar zxf %s ", packName, pkg)
//	kubeLocal := fmt.Sprintf("/root/%s", pkg)
//	var kubeCmd string
//	if packName == "kube" {
//		kubeCmd = "cd /root/kube/&& sh init.sh"
//	} else {
//		kubeCmd = fmt.Sprintf("cd /root/%s && docker load -i images.tar", packName)
//	}
//	var wm sync.WaitGroup
//	for _, host := range hosts {
//		wm.Add(1)
//		go func(host string) {
//			defer wm.Done()
//			logger.Debug("[%s]please wait for decompressing ......", host)
//			if RemoteFilExist(host, kubeLocal) {
//				logger.Warn("[%s]SendPackage: file is exist", host)
//				Cmdout(host, localCmd)
//			} else {
//				if isHttp {
//					go WatchFileSize(host, kubeLocal, GetFileSize(url))
//					Cmdout(host, remoteCmd)
//				} else {
//					Copy(host, url, kubeLocal)
//					Cmdout(host, localCmd)
//				}
//			}
//			Cmdout(host, kubeCmd)
//		}(host)
//	}
//	wm.Wait()
//}

//func KubeadmConfigInstall(){
//	var templateData string
//	if  KubeadmFile == ""{
//		templateData =string(Template())
//
//	}else {
//		fileData, err := ioutil.ReadFile(KubeadmFile)
//		defer func() {
//			if r := recover(); r != nil {
//				logger.Error("[globals]template file read failed:", err)
//			}
//		}()
//		if err != nil {
//			panic(1)
//		}
//		templateData = string(TemplateFromTemplateContent(string(fileData)))
//	}
//	cmd := "echo \"" + templateData + "\" > /root/kube/conf/kubeadm-config.yaml"
//	Cmdout(Masters[0], cmd)
//}

//func LvsInstall(node string){
//	var templateData string
//	if  LvsFile == ""{
//		templateData =string(Template2())
//
//	}else {
//		fileData, err := ioutil.ReadFile(LvsFile)
//		defer func() {
//			if r := recover(); r != nil {
//				logger.Error("[globals]template file read failed:", err)
//			}
//		}()
//		if err != nil {
//			panic(1)
//		}
//		templateData = string(TemplateFromTemplateContent(string(fileData)))
//	}
//
//	cmd := "mkdir -p /etc/kubernetes/manifests/ && echo \"" + templateData + "\" > /etc/kubernetes/manifests/kubeprince-lvs.yaml  "
//	Cmdout(node, cmd)
//
//}
//
//func Lvscreate(node string){
//	var masters string
//	for _, master := range Masters {
//		masters += fmt.Sprintf(" --rs %s:6443", IpFormat(master))
//	}
//	cmd:=fmt.Sprintf("docker run   --rm  -it    --network  host   --privileged=true     fanux/lvscare:latest      /bin/lvscare  create    --vs %s:6443",VIP)
//	cmd+=masters
//	Cmdout(node,cmd)
//}

func ipToInt(ip net.IP) *big.Int {
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}

func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

func stringToIP(i string) net.IP {
	return net.ParseIP(i).To4()
}

// NextIP returns IP incremented by 1
func NextIP(ip net.IP) net.IP {
	i := ipToInt(ip)
	return intToIP(i.Add(i, big.NewInt(1)))
}

// Cmp compares two IPs, returning the usual ordering:
// a < b : -1
// a == b : 0
// a > b : 1
func Cmp(a, b net.IP) int {
	aa := ipToInt(a)
	bb := ipToInt(b)

	if aa == nil || bb == nil {
		logger.Error("ip range %s-%s is invalid", a.String(), b.String())
		os.Exit(-1)
	}
	return aa.Cmp(bb)
}

// ParseIPs 解析ip 192.168.0.2-192.168.0.6
func ParseIPs(ips []string) []string {
	return DecodeIPs(ips)
}

func DecodeIPs(ips []string) []string {
	var res []string
	var port string
	for _, ip := range ips {
		port = "22"
		if ipport := strings.Split(ip, ":"); len(ipport) == 2 {
			ip = ipport[0]
			port = ipport[1]
		}
		if iprange := strings.Split(ip, "-"); len(iprange) == 2 {
			for Cmp(stringToIP(iprange[0]), stringToIP(iprange[1])) <= 0 {
				res = append(res, fmt.Sprintf("%s:%s", iprange[0], port))
				iprange[0] = NextIP(stringToIP(iprange[0])).String()
			}
		} else {
			if stringToIP(ip) == nil {
				logger.Error("ip [%s] is invalid", ip)
				os.Exit(1)
			}
			res = append(res, fmt.Sprintf("%s:%s", ip, port))
		}
	}
	return res
}

// GetMajorMinorInt
func GetMajorMinorInt(version string) (major, minor int) {
	// alpha beta rc version
	if strings.Contains(version, "-") {
		v := strings.Split(version, "-")[0]
		version = v
	}
	version = strings.Replace(version, "v", "", -1)
	versionArr := strings.Split(version, ".")
	if len(versionArr) >= 2 {
		majorStr := versionArr[0] + versionArr[1]
		minorStr := versionArr[2]
		if major, err := strconv.Atoi(majorStr); err == nil {
			if minor, err := strconv.Atoi(minorStr); err == nil {
				return major, minor
			}
		}
	}
	return 0, 0
}

func CanUpgradeByNewVersion(new, old string) error {
	newMajor, newMinor := GetMajorMinorInt(new)
	major, minor := GetMajorMinorInt(old)

	// sealos change cri to containerd when version more than 1.20.0
	if newMajor == 120 && major == 119 {
		return fmt.Errorf("sealos change cri to containerd when Version greater than 1.20! New version: %s, current version: %s", new, old)
	}
	// case one:  new major version <  old major version
	// 1.18.8     1.19.1
	if newMajor < major {
		return fmt.Errorf("kubernetes new version is lower than current version! New version: %s, current version: %s", new, old)
	}
	// case two:  new major version = old major version ; new minor version <= old minor version
	// 1.18.0   1.18.1
	if newMajor == major && newMinor <= minor {
		return fmt.Errorf("kubernetes new version is lower/equal than current version! New version: %s, current version: %s", new, old)
	}

	// case three : new major version > old major version +1;
	// 1.18.2    1.16.10
	if newMajor > major+1 {
		return fmt.Errorf("kubernetes new version is bigger than current version, more than one major version is not allowed! New version: %s, current version: %s", new, old)
	}
	return nil
}

func For120(version string) bool {
	newMajor, _ := GetMajorMinorInt(version)
	// // kubernetes gt 1.20, use Containerd instead of docker
	if newMajor >= 120 {
		logger.Info("install version is: %s, Use kubeadm v1beta2 InitConfig,OCI use containerd instead", version)
		return true
	} else {
		//logger.Info("install version is: %s, Use kubeadm v1beta1 InitConfig, docker", version)
		return false
	}

}

// like y|yes|Y|YES return true
func GetConfirmResult(str string) bool {
	return YesRx.MatchString(str)
}


// send the prompt and get result
func Confirm(prompt string) bool {
	var (
		inputStr string
		err      error
	)
	_, err = fmt.Fprint(os.Stdout, prompt)
	if err != nil {
		logger.Error("fmt.Fprint err", err)
		os.Exit(-1)
	}

	_, err = fmt.Scanf("%s", &inputStr)
	if err != nil {
		logger.Error("fmt.Scanf err", err)
		os.Exit(-1)
	}

	return GetConfirmResult(inputStr)
}

func SliceRemoveStr(ss []string, s string) (result []string) {
	for _, v := range ss {
		if v != s {
			result = append(result, v)
		}
	}
	return
}

//判断当前host的hostname
func isHostName(master, host string) string {
	hostString := SSHConfig.CmdToString(master, "kubectl get nodes | grep -v NAME  | awk '{print $1}'", ",")
	hostName := SSHConfig.CmdToString(host, "hostname", "")
	logger.Debug("hosts %v", hostString)
	hosts := strings.Split(hostString, ",")
	var name string
	for _, h := range hosts {
		if strings.TrimSpace(h) == "" {
			continue
		} else {
			hh := strings.ToLower(h)
			fromH := strings.ToLower(hostName)
			if hh == fromH {
				name = h
				break
			}
		}
	}
	return name
}
