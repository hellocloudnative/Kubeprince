package install

import (
	"Kubeprince/cert"
	"Kubeprince/ipvs"
	"Kubeprince/k8s"
	"Kubeprince/net"
	"bufio"
	"fmt"
	"github.com/wonderivan/logger"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"encoding/json"
)

//SendPackage
func (x *PrinceInstaller) SendPackage() {
	pkg := path.Base(PkgUrl)
	kubeHook := fmt.Sprintf("cd /root && rm -rf kube && tar zxvf %s  && cd /root/kube/shell && rm -f ../bin/kubeprince && bash init-%s.sh", pkg,Containers)
	deletekubectl := `sed -i '/kubectl/d;/kubeprince/d' /root/.bashrc `
	completion := "echo 'command -v kubectl &>/dev/null && source <(kubectl completion bash)' >> /root/.bashrc && echo '[ -x /usr/bin/kubeprince ] && source <(kubeprince completion bash)' >> /root/.bashrc && source /root/.bashrc"
	kubeHook = kubeHook + " && " + deletekubectl + " && " + completion
	SendPackage(PkgUrl, x.Hosts,"/root",nil, &kubeHook)
}

//传包
// SendPackage is send new pkg to all nodes.
func (x *KubeprinceUpgrade) SendPackage() {
	all := append(x.Masters, x.Nodes...)
	pkg := path.Base(x.NewPkgUrl)
	var kubeHook string
	//if For120(Version) {
	//	kubeHook = fmt.Sprintf("cd /root && rm -rf kube && tar zxvf %s  && cd /root/kube/shell && rm -f ../bin/kubeprince && (ctr -n=k8s.io image import ../images/images.tar || true) && cp -f ../bin/* /usr/bin/ ", pkg)
	//} else {
	//	kubeHook = fmt.Sprintf("cd /root && rm -rf kube && tar zxvf %s  && cd /root/kube/shell && rm -f ../bin/kubeprince && (docker load -i ../images/images.tar || true) && cp -f ../bin/* /usr/bin/ ", pkg)
	//
	//}
	if Containers=="isulad"{
		kubeHook = fmt.Sprintf("cd /root && rm -rf kube && tar zxvf %s  && cd /root/kube/shell && rm -f ../bin/kubeprince && (for image_name in $(ls ../images/);do isula  load   -i  ../images/${image_name};done) && cp -f ../bin/* /usr/bin/ ", pkg)
	}else if Containers=="docker"{
		kubeHook = fmt.Sprintf("cd /root && rm -rf kube && tar zxvf %s  && cd /root/kube/shell && rm -f ../bin/kubeprince && (for image_name in $(ls ../images/);do docker  load   -i  ../images/${image_name};done) && cp -f ../bin/* /usr/bin/ ", pkg)

	}

	PkgUrl = SendPackage(pkg, all, "/root", nil, &kubeHook)
}

// SendKubeprince  is send the exec kubeprince to /usr/bin/kubeprince
func (x *PrinceInstaller) Sendkubeprince() {
	// send kubeprince first to avoid old version
	kubeprince := FetchubePrinceAbsPath()
	beforeHook := "ps -ef |grep -v 'grep'|grep kubeprince >/dev/null || rm -rf /usr/bin/kubeprince"
	afterHook := "chmod a+x /usr/bin/kubeprince"
	SendPackage(kubeprince, x.Hosts, "/usr/bin", &beforeHook, &afterHook)
}

func (x *PrinceInstaller) getCgroupDriverFromShell(h string) string {
	var output string
	if For120(Version) {
		cmd := ContainerdShell
		output = SSHConfig.CmdToString(h, cmd, " ")
	} else {
		cmd := DockerShell
		output = SSHConfig.CmdToString(h, cmd, " ")
	}
	output = strings.TrimSpace(output)
	logger.Info("cgroup driver is %s", output)
	return output
}


//KubeadmConfigInstall
func (x *PrinceInstaller)  KubeadmConfigInstall() {
	var templateData string
	CgroupDriver = x.getCgroupDriverFromShell(x.Masters[0])
	if KubeadmFile == "" {
			templateData = string(Template())
	} else {
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
	cmd := fmt.Sprintf(`echo "%s" > /root/kubeadm-config.yaml`, templateData)
	//cmd := "echo \"" + templateData + "\" > /root/kubeadm-config.yaml"
	_ = SSHConfig.CmdAsync(x.Masters[0], cmd)
	//读取模板数据
	kubeadm := KubeadmDataFromYaml(templateData)
	if kubeadm != nil {
		DnsDomain = kubeadm.Networking.DnsDomain
		ApiServerCertSANs = kubeadm.ApiServer.CertSANs
	} else {
		logger.Warn("decode certSANs from config failed, using default SANs")
		ApiServerCertSANs = getDefaultSANs()
	}
}

func getDefaultSANs() []string {
	var sans = []string{"127.0.0.1", "apiserver.cluster.local", VIP}
	// 指定的certSANS不为空, 则添加进去
	if len(CertSANS) != 0 {
		sans = append(sans, CertSANS...)
	}
	for _, master := range Masters {
		sans = append(sans, IpFormat(master))
	}
	return sans
}

//func (p *PrinceInstaller)Command(version string,name CommandType)(cmd string){
//	cmds :=make(map[CommandType]string)
//	cmds= map[CommandType]string{
//		InitMaster: `kubeadm init --config=/root/kube/conf/kubeadm-config.yaml --experimental-upload-certs`,
//		JoinMaster: fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --experimental-control-plane --certificate-key %s", IpFormat(Masters[0]), JoinToken, TokenCaCertHash, CertificateKey),
//		JoinNode:   fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s", VIP, JoinToken, TokenCaCertHash),
//
//	}
//	//other version todo
//	if VersionToInt(version) >= 115{
//		cmds[InitMaster] = `kubeadm init --config=/root/kube/conf/kubeadm-config.yaml --upload-certs`
//		cmds[JoinMaster] = fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --control-plane --certificate-key %s", ApiServer, JoinToken, TokenCaCertHash, CertificateKey)
//	}
//
//	v,ok :=cmds[name]
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Error("[globals]fetch command error")
//		}
//	}()
//	if !ok {
//		panic(1)
//	}
//	return v
//
//
//}
func (x *PrinceInstaller) to11911192(masters []string) (to11911192 bool) {
	// fix > 1.19.1 kube-controller-manager and kube-scheduler use the LocalAPIEndpoint instead of the ControlPlaneEndpoint.
	if VersionToIntAll(Version) >= 1191 && VersionToIntAll(Version) <= 1192 {
		for _, v := range masters {
			ip := IpFormat(v)
			// use grep -qF if already use sed then skip....
			cmd := fmt.Sprintf(`grep -qF "apiserver.cluster.local" %s  && \
sed -i 's/apiserver.cluster.local/%s/' %s && \
sed -i 's/apiserver.cluster.local/%s/' %s`, KUBESCHEDULERCONFIGFILE, ip, KUBECONTROLLERCONFIGFILE, ip, KUBESCHEDULERCONFIGFILE)
			SSHConfig.CmdAsync(v, cmd)
		}
		to11911192 = true
	} else {
		to11911192 = false
	}
	return
}

func (x *PrinceInstaller) sendNewCertAndKey(Hosts []string) {
	var wg sync.WaitGroup
	for _, node := range Hosts {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			SSHConfig.CopyLocalToRemote(node, CertPath, cert.KubeDefaultCertPath)
		}(node)
	}
	wg.Wait()
}

//SendKubeConfigs
func (x *PrinceInstaller) SendKubeConfigs(masters []string) {
	x.sendKubeConfigFile(masters, "kubelet.conf")
	x.sendKubeConfigFile(masters, "admin.conf")
	x.sendKubeConfigFile(masters, "controller-manager.conf")
	x.sendKubeConfigFile(masters, "scheduler.conf")

	if x.to11911192(masters) {
		logger.Info("set 1191 1192 config")
	}
}

func (x *PrinceInstaller) appendApiServer() error {
	etcHostPath := "/etc/hosts"
	etcHostMap := fmt.Sprintf("%s %s", IpFormat(x.Masters[0]), ApiServer)
	file, err := os.OpenFile(etcHostPath, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		os.Exit(1)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		str, err := reader.ReadString('\n')
		if strings.Contains(str, ApiServer) {
			logger.Info("local %s is already exists %s", etcHostPath, ApiServer)
			return nil
		}
		if err == io.EOF {
			break
		}
	}
	write := bufio.NewWriter(file)
	write.WriteString(etcHostMap)
	return write.Flush()
}

//func (p *PrinceInstaller) GeneratorToken() {
//	cmd := `kubeadm token create --print-join-command`
//	output := Cmdout(Masters[0], cmd)
//	decodeOutput(output)
//}

func (x *PrinceInstaller) GeneratorCert() {
	//cert generator in kubeprince
	hostname := GetRemoteHostName(x.Masters[0])
	cert.GenerateCert(CertPath, CertEtcdPath, ApiServerCertSANs, IpFormat(x.Masters[0]), hostname, SvcCIDR, DnsDomain)
	//copy all cert to master0
	//CertSA(kye,pub) + CertCA(key,crt)
	//s.sendNewCertAndKey(s.Masters)
	//s.sendCerts([]string{s.Masters[0]})
}

func (x *PrinceInstaller) CreateKubeconfig() {
	hostname := GetRemoteHostName(x.Masters[0])

	certConfig := cert.Config{
		Path:     CertPath,
		BaseName: "ca",
	}

	controlPlaneEndpoint := fmt.Sprintf("https://%s:6443", ApiServer)

	err := cert.CreateJoinControlPlaneKubeConfigFiles(cert.KubeprinceConfigDir,
		certConfig, hostname, controlPlaneEndpoint, "kubernetes")
	if err != nil {
		logger.Error("generator kubeconfig failed %s", err)
		os.Exit(-1)
	}

}

func (x *PrinceInstaller) SendJoinMasterKubeConfigs(masters []string) {
	x.sendKubeConfigFile(masters, "admin.conf")
	x.sendKubeConfigFile(masters, "controller-manager.conf")
	x.sendKubeConfigFile(masters, "scheduler.conf")
	if x.to11911192(masters) {
		logger.Info("set 1191 1192 config")
	}
}

func (x *PrinceInstaller)InstallMaster0(){
	x.SendKubeConfigs([]string{x.Masters[0]})
	x.sendNewCertAndKey([]string{x.Masters[0]})

	// remote server run kubeprince init . it can not reach apiserver.cluster.local , should add masterip apiserver.cluster.local to /etc/hosts
	err := x.appendApiServer()
	if err != nil {
		logger.Warn("append  %s %s to /etc/hosts err: %s", IpFormat(x.Masters[0]), ApiServer, err)
	}

	//master0 do sth
	cmd := fmt.Sprintf("grep -qF '%s %s' /etc/hosts || echo %s %s >> /etc/hosts", IpFormat(x.Masters[0]), ApiServer, IpFormat(x.Masters[0]), ApiServer)
	_ = SSHConfig.CmdAsync(x.Masters[0], cmd)

	cmd = x.Command(Version, InitMaster)

	output := SSHConfig.Cmd(x.Masters[0], cmd)
	if output == nil {
		logger.Error("[%s] install kubernetes failed. please clean and uninstall.", x.Masters[0])
		os.Exit(1)
	}
	decodeOutput(output)

	cmd = `mkdir -p /root/.kube && cp /etc/kubernetes/admin.conf /root/.kube/config && chmod 600 /root/.kube/config`
	output = SSHConfig.Cmd(x.Masters[0], cmd)

	if WithoutCNI {
		logger.Info("--without-cni is true, so we not install calico or flannel, install it by yourself")
		return
	}
	//cmd = `kubectl apply -f /root/kube/conf/net/calico.yaml || true`

	// can-reach is used by calico multi network , flannel has nothing to add. just Use it.
	if k8s.IsIpv4(Interface) && Network == "calico" {
		Interface = "can-reach=" + Interface
	} else if !k8s.IsIpv4(Interface) && Network == "calico"  {
		Interface = "interface=" + Interface
	}

	var cniVersion string
	if SSHConfig.IsFileExist(x.Masters[0], "/root/kube/Metadata") {
		var metajson string
		var tmpdata metadata
		metajson = SSHConfig.CmdToString(x.Masters[0], "cat /root/kube/Metadata", "")
		err := json.Unmarshal([]byte(metajson), &tmpdata)
		if err != nil {
			logger.Warn("get metadata version err: ", err)
		} else {
			cniVersion = tmpdata.CniVersion
			Network = tmpdata.CniName
		}
	}

	netyaml := net.NewNetwork(Network, net.MetaData{
		Interface:      Interface,
		CIDR:           PodCIDR,
		IPIP:           !BGP,
		MTU:            MTU,
		CniRepo:        Repo,
		K8sServiceHost: x.ApiServer,
		Version:        cniVersion,
	}).Manifests("")
	logger.Debug("cni yaml : \n", netyaml)
	home := cert.GetUserHomeDir()
	configYamlDir := filepath.Join(home, ".kubeprince", "cni.yaml")
	ioutil.WriteFile(configYamlDir, []byte(netyaml), 0755)
	SSHConfig.Copy(x.Masters[0], configYamlDir, "/tmp/cni.yaml")
	output = SSHConfig.Cmd(x.Masters[0], "kubectl apply -f /tmp/cni.yaml")
}

//join master
func (x *PrinceInstaller) JoinMasters(masters []string) {
	var wg sync.WaitGroup
	//copy certs & kube-config
	x.SendJoinMasterKubeConfigs(masters)
	x.sendNewCertAndKey(masters)
	// send CP nodes configuration
	x.sendJoinCPConfig(masters)

	//join master do sth
	cmd := x.Command(Version, JoinMaster)
	for _, master := range masters {
		wg.Add(1)
		go func(master string) {
			defer wg.Done()
			hostname := GetRemoteHostName(master)
			certCMD := cert.CMD(ApiServerCertSANs, IpFormat(master), hostname, SvcCIDR, DnsDomain)
			_ = SSHConfig.CmdAsync(master, certCMD)

			cmdHosts := fmt.Sprintf("echo %s >> /etc/hosts", getApiserverHost(IpFormat(x.Masters[0])))
			_ = SSHConfig.CmdAsync(master, cmdHosts)
			// cmdMult := fmt.Sprintf("%s --apiserver-advertise-address %s", cmd, IpFormat(master))
			_ = SSHConfig.CmdAsync(master, cmd)
			cmdHosts = fmt.Sprintf(`sed "s/%s/%s/g" -i /etc/hosts`, getApiserverHost(IpFormat(x.Masters[0])), getApiserverHost(IpFormat(master)))
			_ = SSHConfig.CmdAsync(master, cmdHosts)
			copyk8sConf := `rm -rf .kube/config && mkdir -p /root/.kube && cp /etc/kubernetes/admin.conf /root/.kube/config && chmod 600 /root/.kube/config`
			_ = SSHConfig.CmdAsync(master, copyk8sConf)
			cleaninstall := `rm -rf /root/kube || :`
			_ = SSHConfig.CmdAsync(master, cleaninstall)
		}(master)
	}
	wg.Wait()
}

// sendJoinCPConfig send join CP nodes configuration
func (x *PrinceInstaller)sendJoinCPConfig(joinMaster []string) {
	var wg sync.WaitGroup
	for _, master := range joinMaster {
		wg.Add(1)
		go func(master string) {
			defer wg.Done()
			cgroup := x.getCgroupDriverFromShell(master)
			templateData := string(JoinTemplate(IpFormat(master),cgroup))
			cmd := fmt.Sprintf(`echo "%s" > /root/kubeadm-join-config.yaml`, templateData)
			_ = SSHConfig.CmdAsync(master, cmd)
		}(master)
	}
	wg.Wait()
}

//join node
func (x *PrinceInstaller) JoinNodes() {
	var masters string
	var wg sync.WaitGroup
	for _, master := range x.Masters {
		masters += fmt.Sprintf(" --rs %s:6443", IpFormat(master))
	}
	ipvsCmd := fmt.Sprintf("kubeprince ipvs --vs %s:6443 %s --health-path /healthz --health-schem https --run-once", VIP, masters)
	for _, node := range x.Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			cgroup := x.getCgroupDriverFromShell(node)
			templateData := string(JoinTemplate("", cgroup))
			// send join node config
			cmdJoinConfig := fmt.Sprintf(`echo "%s" > /root/kubeadm-join-config.yaml`, templateData)
			_ = SSHConfig.CmdAsync(node, cmdJoinConfig)

			cmdHosts := fmt.Sprintf("echo %s %s >> /etc/hosts", VIP, ApiServer)
			_ = SSHConfig.CmdAsync(node, cmdHosts)

			// 如果不是默认路由， 则添加 vip 到 master的路由。
			cmdRoute := fmt.Sprintf("kubeprince route --host %s", IpFormat(node))
			status := SSHConfig.CmdToString(node, cmdRoute, "")
			if status != "ok" {
				// 以自己的ip作为路由网关
				addRouteCmd := fmt.Sprintf("kubeprince route add --host %s --gateway %s", VIP, IpFormat(node))
				SSHConfig.CmdToString(node, addRouteCmd, "")
			}

			_ = SSHConfig.CmdAsync(node, ipvsCmd) // create ipvs rules before we join node
			cmd := x.Command(Version, JoinNode)
			//create lvsucc static pod
			yaml := ipvs.LvsStaticPodYaml(VIP, Masters, LvsuccImage)
			_ = SSHConfig.CmdAsync(node, cmd)
			_ = SSHConfig.Cmd(node, "mkdir -p /etc/kubernetes/manifests")
			SSHConfig.CopyConfigFile(node, "/etc/kubernetes/manifests/kube-lvsucc.yaml", []byte(yaml))

			cleaninstall := `rm -rf /root/kube`
			_ = SSHConfig.CmdAsync(node, cleaninstall)
		}(node)
	}

	wg.Wait()
}

//join master
//func (p *PrinceInstaller) JoinMasters() {
//	for _, master := range Masters[1:] {
//		cmd := p.Command(Version, JoinMaster)
//		logger.Info("[%s]", master)
//		cmdHosts := fmt.Sprintf("echo %s %s >> /etc/hosts", IpFormat(Masters[0]), ApiServer)
//		Cmdout(master, cmdHosts)
//		Cmd(master, cmd)
//		cmdHosts = fmt.Sprintf(`sed "s/%s/%s/g" -i /etc/hosts`, IpFormat(Masters[0]), IpFormat(master))
//		Cmdout(master, cmdHosts)
//		cmd = `mkdir -p /root/.kube && cp /etc/kubernetes/admin.conf /root/.kube/config`
//		Cmdout(master, cmd)
//
//	}
//}

//join node
//func (p *PrinceInstaller) JoinNodes() {
//	var masters string
//	var wg sync.WaitGroup
//	var cmdHosts string
//	for _, master := range Masters {
//		masters += fmt.Sprintf(" --master %s:6443", IpFormat(master))
//	}
//
//	for _, node := range Nodes {
//		wg.Add(1)
//		go func(node string) {
//			defer wg.Done()
//			if len(Masters)>=0{
//				cmdHosts = fmt.Sprintf("echo %s %s >> /etc/hosts", VIP, ApiServer)
//			}else {
//				cmdHosts = fmt.Sprintf("echo %s %s >> /etc/hosts", IpFormat(Masters[0]), ApiServer)
//			}
//
//			Cmdout(node, cmdHosts)
//			Lvscreate(node)
//			cmd := p.Command(Version, JoinNode)
//			Cmdout(node, cmd)
//			LvsInstall(node)
//
//		}(node)
//	}
//
//	wg.Wait()
//}



