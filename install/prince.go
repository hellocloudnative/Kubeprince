package install

import (
	"fmt"
	"github.com/wonderivan/logger"
	"os"
	"sync"
)

//SendPackage
func (p *PrinceInstaller) SendPackage(packName string) {
	SendPackage(PkgUrl, p.Hosts, packName)
}
//KubeadmConfigInstall
func (p *PrinceInstaller) KubeadmConfigInstall(){
	KubeadmConfigInstall()
}

func (p *PrinceInstaller)Command(version string,name CommandType)(cmd string){
	cmds :=make(map[CommandType]string)
	cmds= map[CommandType]string{
		InitMaster: `kubeadm init --config=/root/kube/conf/kubeadm-config.yaml --experimental-upload-certs`,
		JoinMaster: fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --experimental-control-plane --certificate-key %s", IpFormat(Masters[0]), JoinToken, TokenCaCertHash, CertificateKey),
		JoinNode:   fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s", IpFormat(Masters[0]), JoinToken, TokenCaCertHash),

	}
	//other version todo
	if VersionToInt(version) >= 115{
		cmds[InitMaster] = `kubeadm init --config=/root/kube/conf/kubeadm-config.yaml --upload-certs`
		cmds[JoinMaster] = fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --control-plane --certificate-key %s", IpFormat(Masters[0]), JoinToken, TokenCaCertHash, CertificateKey)
	}
	if len(Masters)>=3{
		cmds[JoinNode]= fmt.Sprintf("kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s", VIP, JoinToken, TokenCaCertHash)
	}
	v,ok :=cmds[name]
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
func (p *PrinceInstaller)InstallMaster0(){
	cmd := fmt.Sprintf("echo %s %s >> /etc/hosts", IpFormat(Masters[0]), ApiServer)
	Cmd(Masters[0], cmd)
	cmd = p.Command(Version,InitMaster)
	output:=Cmd(Masters[0],cmd)
	if output == nil {
		logger.Error("[%s]kubernetes install is error.please clean and uninstall.", Masters[0])
		os.Exit(1)
	}
	decodeOutput(output)
	cmd = `mkdir -p /root/.kube && cp /etc/kubernetes/admin.conf /root/.kube/config`
	output = Cmd(Masters[0], cmd)

	cmd = `kubectl apply -f /root/kube/conf/calico.yaml || true`
	output = Cmd(Masters[0], cmd)
}

func (p *PrinceInstaller) GeneratorToken() {
	cmd := `kubeadm token create --print-join-command`
	output := Cmd(Masters[0], cmd)
	decodeOutput(output)
}
//join master
func (p *PrinceInstaller) JoinMasters() {
	cmd := p.Command(Version, JoinMaster)
	for _, master := range Masters[1:] {
		cmdHosts := fmt.Sprintf("echo %s %s >> /etc/hosts", IpFormat(Masters[0]), ApiServer)
		Cmd(master, cmdHosts)
		Cmd(master, cmd)
		cmdHosts = fmt.Sprintf(`sed "s/%s/%s/g" -i /etc/hosts`, IpFormat(Masters[0]), IpFormat(master))
		Cmd(master, cmdHosts)
		cmd = `mkdir -p /root/.kube && cp /etc/kubernetes/admin.conf /root/.kube/config`
		Cmd(master, cmd)

	}
}
//join node
func (p *PrinceInstaller) JoinNodes() {
	//var masters string
	var wg sync.WaitGroup
	var cmdHosts string
	//for _, master := range Masters {
	//	masters += fmt.Sprintf(" --master %s:6443", IpFormat(master))
	//}

	for _, node := range Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			if len(Masters)>=3{
				cmdHosts = fmt.Sprintf("echo %s %s >> /etc/hosts", VIP, ApiServer)
			}else {
				cmdHosts = fmt.Sprintf("echo %s %s >> /etc/hosts", IpFormat(Masters[0]), ApiServer)
			}

			Cmd(node, cmdHosts)
			cmd := p.Command(Version, JoinNode)
			//cmd += masters
			Cmd(node, cmd)

		}(node)
	}

	wg.Wait()
}


//CleanCluster
func (p *PrinceInstaller)Clean() {
	var wg sync.WaitGroup
	for _, host := range p.Hosts {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			clean(node)
		}(host)
	}
	wg.Wait()
}
func clean(host string) {
	cmd := "kubeadm reset -f && modprobe -r ipip  && lsmod"
	Cmd(host, cmd)
	cmd = "rm -rf ~/.kube/ && rm -rf /etc/kubernetes/"
	Cmd(host, cmd)
	cmd = "rm -rf /etc/systemd/system/kubelet.service.d && rm -rf /etc/systemd/system/kubelet.service"
	Cmd(host, cmd)
	cmd = "rm -rf /usr/bin/kube* && rm -rf /usr/bin/crictl"
	Cmd(host, cmd)
	cmd = "rm -rf /etc/cni && rm -rf /opt/cni"
	Cmd(host, cmd)
	cmd = "rm -rf /var/lib/etcd && rm -rf /var/etcd "
	Cmd(host, cmd)
	cmd = fmt.Sprintf("rm -rf /tmp/* && sed -i \"/%s/d\" /etc/hosts ", ApiServer)
	Cmd(host, cmd)
}