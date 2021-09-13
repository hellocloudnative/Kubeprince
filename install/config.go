package install

import (
	"Kubeprince/cert"
	"Kubeprince/ipvs"
	"Kubeprince/pkg/sshcmd/sshutil"
	"bytes"
	"fmt"
	"github.com/wonderivan/logger"
	"html/template"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"gopkg.in/yaml.v2"
	"lvsucc/care"
)
const defaultKubePath = "/kube"
const defaultConfigPath = "/.kubeprince"
const defaultConfigFile = "/config.yaml"

type PrinceConfig struct {
	Masters []string
	Nodes   []string
	//config from kubeadm.cfg. ex. cluster.local
	DnsDomain         string
	ApiServerCertSANs []string

	//SSHConfig
	User       string
	Passwd     string
	PrivateKey string
	PkPassword string
	//ApiServer ex. apiserver.cluster.local
	ApiServerDomain string
	Network         string
	VIP             string
	PkgURL          string
	Version         string
	Repo            string
	PodCIDR         string
	SvcCIDR         string
	//certs location
	CertPath     string
	CertEtcdPath string
	//lvsucc images
	LvsuccName string
	LvsuccTag  string
}

type PrinceInstaller struct {
	Hosts []string
	Masters   []string
	Nodes     []string
	Network   string
	ApiServer string
}

var (
	Masters  []string
	Nodes    []string
	SSHConfig sshutil.SSH
	CertSANS  []string
	//config from kubeadm.cfg
	DnsDomain         string
	ApiServerCertSANs []string
	VIP            string
	PkgUrl         string
	//User           string
	//Password       string
	//PrivateKeyFile string
	KubeadmFile    string
	//LvsFile        string
	Version        string
	//Kustomize      bool
	ApiServer      string
	CleanForce bool
	CleanAll   bool
	UpdateForce bool
	UpdateAll   bool
	Vlog int
	Repo    string
	PodCIDR string
	SvcCIDR string
	// the ipip mode of the calico
	BGP bool
	// mtu size
	MTU string
	// if true don't install cni plugin
	WithoutCNI bool
	//network interface name, like "eth.*|en.*"
	YesRx = regexp.MustCompile("^(?i:y(?:es)?)$")
	Interface string
	//criSocket
	CriSocket string

	Ipvs         care.LvsCare
	LvsuccImage ipvs.LvsuccImage

	CertPath     = cert.KubeprinceConfigDir + "/pki"
	CertEtcdPath = cert.KubeprinceConfigDir + "/pki/etcd"
	EtcdCacart   = cert.KubeprinceConfigDir + "/pki/etcd/ca.crt"
	EtcdCert     = cert.KubeprinceConfigDir + "/pki/etcd/healthcheck-client.crt"
	EtcdKey      = cert.KubeprinceConfigDir + "/pki/etcd/healthcheck-client.key"
	// network type, calico or flannel etc..
	Network string
	Containers string
)

type metadata struct {
	K8sVersion string `json:"k8sVersion"`
	CniVersion string `json:"cniVersion"`
	CniName    string `json:"cniName"`
}

const (
	ErrorExitOSCase = -1 // 错误直接退出类型

	ErrorMasterEmpty    = "your master is empty."                 // master节点ip为空
	ErrorVersionEmpty   = "your kubernetes version is empty."     // kubernetes 版本号为空
	ErrorFileNotExist   = "your package file is not exist."       // 离线安装包为空
	ErrorPkgUrlNotExist = "Your package url is incorrect."        // 离线安装包为http路径不对
	ErrorPkgUrlSize     = "Download file size is less then 200M " // 离线安装包为http路径不对
	//ErrorMessageSSHConfigEmpty = "your ssh password or private-key is empty."		// ssh 密码/秘钥为空
	// ErrorMessageCommon											// 其他错误消息

	// MinDownloadFileSize int64 = 400 * 1024 * 1024

	// etcd backup
	ETCDSNAPSHOTDEFAULTNAME = "snapshot"
	ETCDDEFAULTBACKUPDIR    = "/opt/sealos/etcd-backup"
	ETCDDEFAULTRESTOREDIR   = "/opt/sealos/etcd-restore"
	ETCDDATADIR             = "/var/lib/etcd"
	TMPDIR                  = "/tmp"

	// kube file
	KUBECONTROLLERCONFIGFILE = "/etc/kubernetes/controller-manager.conf"
	KUBESCHEDULERCONFIGFILE  = "/etc/kubernetes/scheduler.conf"

	// CriSocket
	 DefaultDockerCRISocket     = "/var/run/docker.sock"
	//DefaultContainerdCRISocket = "/run/containerd/containerd.sock"
	DefaultiSuladCRISocket = "/var/run/isulad.sock"
)

const InitTemplateTextV1beta1 = string(`apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: {{.Master0}}
  bindPort: 6443
nodeRegistration:
  criSocket: /var/run/isulad.sock
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
kubernetesVersion: {{.Version}}
controlPlaneEndpoint: "{{.ApiServer}}:6443"
imageRepository: {{.Repo}}
networking:
  # dnsDomain: cluster.local
  podSubnet: {{.PodCIDR}}
  serviceSubnet: {{.SvcCIDR}}
apiServer:
  certSANs:
  - 127.0.0.1
  - {{.ApiServer}}
  {{range .Masters -}}
  - {{.}}
  {{end -}}
  {{range .CertSANS -}}
  - {{.}}
  {{end -}}
  - {{.VIP}}
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - name: localtime
    hostPath: /etc/localtime
    mountPath: /etc/localtime
    readOnly: true
    pathType: File
controllerManager:
  extraArgs:
    feature-gates: TTLAfterFinished=true
    experimental-cluster-signing-duration: 876000h
{{- if eq .Network "cilium" }}
    allocate-node-cidrs: \"true\"
{{- end }}
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File
scheduler:
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File
---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
mode: "ipvs"
ipvs:
  excludeCIDRs:
  - "{{.VIP}}/32"`)

const JoinCPTemplateTextV1beta2 = string(`apiVersion: kubeadm.k8s.io/v1beta2
caCertPath: /etc/kubernetes/pki/ca.crt
discovery:
  bootstrapToken:
    {{- if .Master}}
    apiServerEndpoint: {{.Master0}}:6443
    {{else}}
    apiServerEndpoint: {{.VIP}}:6443
    {{end -}}
    token: {{.TokenDiscovery}}
    caCertHashes:
    - {{.TokenDiscoveryCAHash}}
  timeout: 5m0s
kind: JoinConfiguration
{{- if .Master }}
controlPlane:
  localAPIEndpoint:
    advertiseAddress: {{.Master}}
    bindPort: 6443
{{- end}}
nodeRegistration:
  criSocket: {{.CriSocket}}`)

const InitTemplateTextV1beta2 = string(`apiVersion: kubeadm.k8s.io/v1beta2
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: {{.Master0}}
  bindPort: 6443
---
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
kubernetesVersion: {{.Version}}
controlPlaneEndpoint: "{{.ApiServer}}:6443"
imageRepository: {{.Repo}}
networking:
  # dnsDomain: cluster.local
  podSubnet: {{.PodCIDR}}
  serviceSubnet: {{.SvcCIDR}}
apiServer:
  certSANs:
  - 127.0.0.1
  - {{.ApiServer}}
  {{range .Masters -}}
  - {{.}}
  {{end -}}
  {{range .CertSANS -}}
  - {{.}}
  {{end -}}
  - {{.VIP}}
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - name: localtime
    hostPath: /etc/localtime
    mountPath: /etc/localtime
    readOnly: true
    pathType: File
controllerManager:
  extraArgs:
    feature-gates: TTLAfterFinished=true
    experimental-cluster-signing-duration: 876000h
{{- if eq .Network "cilium" }}
    allocate-node-cidrs: \"true\"
{{- end }}
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File
scheduler:
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File
---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
mode: "ipvs"
ipvs:
  excludeCIDRs:
  - "{{.VIP}}/32"`)


var (
	JoinToken       string
	TokenCaCertHash string
	CertificateKey  string
)

type CommandType string

const InitMaster CommandType = "initMaster"
const JoinMaster CommandType = "joinMaster"
const JoinNode CommandType = "joinNode"

func (p *PrinceInstaller) Command(version string, name CommandType) (cmd string) {
	cmds := make(map[CommandType]string)
	// Please convert your v1beta1 configuration files to v1beta2 using the
	// "kubeadm config migrate" command of kubeadm v1.15.x, 因此1.14 版本不支持双网卡.
	cmds = map[CommandType]string{
		InitMaster: `kubeadm init --config=/root/kubeadm-config.yaml --experimental-upload-certs` + vlogToStr(),
		JoinMaster: fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --experimental-control-plane --certificate-key %s"+vlogToStr(), IpFormat(p.Masters[0]), JoinToken, TokenCaCertHash, CertificateKey),
		JoinNode:   fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s"+vlogToStr(), VIP, JoinToken, TokenCaCertHash),
	}
	//other version >= 1.15.x
	//todo
	if VersionToInt(version) >= 115 {
		cmds[InitMaster] = `kubeadm init --config=/root/kubeadm-config.yaml --upload-certs` + vlogToStr()
		cmds[JoinMaster] = "kubeadm join --config=/root/kubeadm-join-config.yaml " + vlogToStr()
		cmds[JoinNode] = "kubeadm join --config=/root/kubeadm-join-config.yaml " + vlogToStr()
	}


	if p.Network == "cilium" {
		if VersionToInt(version) >= 116 {
			cmds[InitMaster] = `kubeadm init --skip-phases=addon/kube-proxy --config=/root/kubeadm-config.yaml --upload-certs` + vlogToStr()
		} else {
			cmds[InitMaster] = `kubeadm init --config=/root/kubeadm-config.yaml --upload-certs` + vlogToStr()
		}
	}

	v, ok := cmds[name]
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[globals]fetch command error")
		}
	}()
	if !ok {
		panic(1)
	}
	return v
}

func vlogToStr() string {
	str := strconv.Itoa(Vlog)
	return " -v " + str
}

// JoinTemplate is generate JoinCP nodes configuration by master ip.
func JoinTemplate(ip string) []byte {
	return JoinTemplateFromTemplateContent(joinKubeadmConfig(), ip)
}

func joinKubeadmConfig() string {
	var sb strings.Builder
	sb.Write([]byte(JoinCPTemplateTextV1beta2))
	return sb.String()
}

//Load is
func (x *PrinceConfig) Load(path string) (err error) {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = home + defaultConfigPath + defaultConfigFile
	}

	y, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %s failed %w", path, err)
	}

	err = yaml.Unmarshal(y,x)
	if err != nil {
		return fmt.Errorf("unmarshal config file failed: %w", err)
	}

	Masters = x.Masters
	Nodes = x.Nodes
	SSHConfig.User = x.User
	SSHConfig.Password = x.Passwd
	SSHConfig.PkFile = x.PrivateKey
	SSHConfig.PkPassword = x.PkPassword
	ApiServer = x.ApiServerDomain
	Network = x.Network
	VIP = x.VIP
	PkgUrl = x.PkgURL
	Version = x.Version
	Repo = x.Repo
	PodCIDR = x.PodCIDR
	SvcCIDR = x.SvcCIDR
	DnsDomain = x.DnsDomain
	ApiServerCertSANs =x.ApiServerCertSANs
	CertPath = x.CertPath
	CertEtcdPath = x.CertEtcdPath
	//lvscare
	LvsuccImage.Image = x.LvsuccName
	LvsuccImage.Tag = x.LvsuccTag
	return
}


//const Templateyaml= string(`apiVersion: kubeadm.k8s.io/v1beta1
//kind: ClusterConfiguration
//kubernetesVersion: {{.Version}}
//controlPlaneEndpoint: "{{.ApiServer}}:6443"
//networking:
//  podSubnet: 100.64.0.0/10
//apiServer:
//        certSANs:
//        - 127.0.0.1
//        - {{.ApiServer}}
//        {{range .Masters -}}
//        - {{.}}
//        {{end -}}
//        - {{.VIP}}
//---
//apiVersion: kubeproxy.config.k8s.io/v1alpha1
//kind: KubeProxyConfiguration
//mode: "ipvs"
//ipvs:
//        excludeCIDRs:
//        - "{{.VIP}}/32"`)

//var ConfigType string
//func Config() {
//	switch ConfigType {
//	case "kubeadm":
//		printlnKubeadmConfig()
//	default:
//		printlnKubeadmConfig()
//	}
//}
//func printlnKubeadmConfig() {
//	fmt.Println(kubeadmConfig())
//}

//const  princelvsyaml CommandType =(`
//apiVersion: v1
//kind: Pod
//metadata:
//  labels:
//    component: kubeprince-lvs
//    tier: control-plane
//  name: kubeprince-lvs
//  namespace: kube-system
//spec:
//  containers:
//  - command:
//    - /usr/bin/lvscare
//    - care
//    - --vs
//    - {{.VIP}}:6443
//    - --health-path
//    - /healthz
//    - --health-schem
//    - https
//    - --rs
//    {{range .Masters -}}
//    - {{.}}:6443
//    {{end -}}
//    image: kubeprince-lvs:latest
//    imagePullPolicy: IfNotPresent
//    name: kubeprince-lvs
//    securityContext:
//      privileged: true
//  hostNetwork: true
//  priorityClassName: system-cluster-critical
//status: {}`)

//func kubeadmConfig() (string) {
//	var sb strings.Builder
//	sb.Write([]byte(Templateyaml))
//	return sb.String()
//}
var ConfigType string

func Config() {
	switch ConfigType {
	case "kubeadm":
		printlnKubeadmConfig()
	case "join":
		printlnJoinKubeadmConfig()
	default:
		printlnKubeadmConfig()
	}
}

func printlnKubeadmConfig() {
	fmt.Println(kubeadmConfig())
}

func printlnJoinKubeadmConfig() {
	fmt.Println(joinKubeadmConfig())
}

func kubeadmConfig() string {
	var sb strings.Builder
	if Containers=="docker" {
		sb.Write([]byte(InitTemplateTextV1beta2))
	} else if Containers=="isulad" {
		sb.Write([]byte(InitTemplateTextV1beta1))
	}

	return sb.String()
}

//func lvsConfig() (string) {
//	var sb strings.Builder
//	sb.Write([]byte(princelvsyaml))
//	return sb.String()
//}

func  Template()([]byte){
	return  TemplateFromTemplateContent(kubeadmConfig())
}

//func  Template2()([]byte){
//	return  TemplateFromTemplateContent(lvsConfig())
//}

func JoinTemplateFromTemplateContent(templateContent, ip string) []byte {
	tmpl, err := template.New("text").Parse(templateContent)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("join template parse failed:", err)
		}
	}()
	if err != nil {
		panic(1)
	}
	var envMap = make(map[string]interface{})
	envMap["Master0"] = IpFormat(Masters[0])
	envMap["Master"] = ip
	envMap["TokenDiscovery"] = JoinToken
	envMap["TokenDiscoveryCAHash"] = TokenCaCertHash
	envMap["VIP"] = VIP
	//if For120(Version) {
	//	CriSocket = DefaultContainerdCRISocket
	//} else {
	//	CriSocket = DefaultDockerCRISocket
	//}
	if Containers=="isulad"{
		CriSocket = DefaultiSuladCRISocket
	}else if Containers=="docker"{
		CriSocket = DefaultDockerCRISocket
	}

	envMap["CriSocket"] = CriSocket
	var buffer bytes.Buffer
	_ = tmpl.Execute(&buffer, envMap)
	return buffer.Bytes()
}


//func TemplateFromTemplateContent(templateContent string) []byte {
//	tmpl, err := template.New("text").Parse(templateContent)
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("template parse failed:", err)
//		}
//	}()
//	if err != nil {
//		panic(1)
//	}
//	var masters []string
//	for _, h := range Masters {
//		masters = append(masters, IpFormat(h))
//	}
//	var envMap = make(map[string]interface{})
//	envMap["VIP"] = VIP
//	envMap["Masters"] = masters
//	envMap["Version"] = Version
//	envMap["ApiServer"] = ApiServer
//	var buffer bytes.Buffer
//	_ = tmpl.Execute(&buffer, envMap)
//	return buffer.Bytes()
//}
func TemplateFromTemplateContent(templateContent string) []byte {
	tmpl, err := template.New("text").Parse(templateContent)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("template parse failed:", err)
		}
	}()
	if err != nil {
		panic(1)
	}
	var masters []string
	getmasters := Masters
	for _, h := range getmasters {
		masters = append(masters, IpFormat(h))
	}
	var envMap = make(map[string]interface{})
	envMap["CertSANS"] = CertSANS
	envMap["VIP"] = VIP
	envMap["Masters"] = masters
	envMap["Version"] = Version
	envMap["ApiServer"] = ApiServer
	envMap["PodCIDR"] = PodCIDR
	envMap["SvcCIDR"] = SvcCIDR
	envMap["Repo"] = Repo
	envMap["Master0"] = IpFormat(Masters[0])
	envMap["Network"] = Network
	var buffer bytes.Buffer
	_ = tmpl.Execute(&buffer, envMap)
	return buffer.Bytes()
}

func (x *PrinceConfig) ShowDefaultConfig() {
	home, _ := os.UserHomeDir()
	x.Masters = []string{"192.168.0.2", "192.168.0.2", "192.168.0.2"}
	x.Nodes = []string{"192.168.0.3", "192.168.0.4"}
	x.User = "root"
	x.Passwd = "123456"
	x.PrivateKey = home + "/.ssh/id_rsa"
	x.ApiServerDomain = "apiserver.cluster.local"
	x.Network = "calico"
	x.VIP = "10.103.97.2"
	x.PkgURL = home + "/kube1.17.13.tar.gz"
	x.Version = "v1.17.13"
	x.Repo = "k8s.gcr.io"
	x.PodCIDR = "100.64.0.0/10"
	x.SvcCIDR = "10.96.0.0/12"
	x.ApiServerCertSANs = []string{"apiserver.cluster.local", "127.0.0.1"}
	x.CertPath = home + "/.kubeprince/pki"
	x.CertEtcdPath = home + "/.kubeprince/pki/etcd"
	x.LvsuccName = "harbor.sh.deepin.com/lvsucc"
	x.LvsuccTag = "latest"

	y, err := yaml.Marshal(x)
	if err != nil {
		logger.Error("marshal config file failed: %s", err)
	}

	logger.Info("\n\n%s\n\n", string(y))
	logger.Info("Please save above config in ~/.kubeprince/config.yaml and edit values on your own")
}

//Dump is
func (x *PrinceConfig) Dump(path string) {
	home, _ := os.UserHomeDir()
	if path == "" {
		path = home + defaultConfigPath + defaultConfigFile
	}
	Masters = ParseIPs(Masters)
	x.Masters = Masters
	Nodes = ParseIPs(Nodes)
	x.Nodes = ParseIPs(Nodes)
	x.User = SSHConfig.User
	x.Passwd = SSHConfig.Password
	x.PrivateKey = SSHConfig.PkFile
	x.PkPassword = SSHConfig.PkPassword
	x.ApiServerDomain = ApiServer
	x.Network = Network
	x.VIP = VIP
	x.PkgURL = PkgUrl
	x.Version = Version
	x.Repo = Repo
	x.SvcCIDR = SvcCIDR
	x.PodCIDR = PodCIDR

	x.DnsDomain = DnsDomain
	x.ApiServerCertSANs = ApiServerCertSANs
	x.CertPath = CertPath
	x.CertEtcdPath = CertEtcdPath
	//lvscare
	x.LvsuccName = LvsuccImage.Image
	x.LvsuccTag = LvsuccImage.Tag

	y, err := yaml.Marshal(x)
	if err != nil {
		logger.Error("dump config file failed: %s", err)
	}


	err = os.MkdirAll(home+defaultConfigPath, os.ModePerm)
	if err != nil {
		logger.Warn("create default kubeprince  config dir failed, please create it by your self mkdir -p /root/.kubeprince && touch /root/.kubeprince /config.yaml")
	}

	if err = ioutil.WriteFile(path, y, 0644); err != nil {
		logger.Warn("write to file %s failed: %s", path, err)
	}
}

