package install

import (
	"Kubeprince/ipvs"
	"fmt"
	"github.com/wonderivan/logger"
	"os"
	"strings"
	sshcmd "Kubeprince/pkg/sshcmd/cmd"
	"sync"
)
type PrinceClean struct {
	PrinceInstaller
	cleanAll bool
}

// BuildClean clean the build resources.
func BuildClean(deleteNodes, deleteMasters []string) {
	i := &PrinceClean{cleanAll: false}
	masters := Masters
	nodes := Nodes
	//1. 删除masters
	if len(deleteMasters) != 0 {
		if !CleanForce { // false
			prompt := fmt.Sprintf("clean command will clean masters [%s], continue clean (y/n)?", strings.Join(deleteMasters, ","))
			result := Confirm(prompt)
			if !result {
				logger.Debug("clean masters command is skip")
				goto node
			}
		}
		//只清除masters
		i.Masters = deleteMasters
	}

	//2. 删除nodes
node:
	if len(deleteNodes) != 0 {
		if !CleanForce { // flase
			prompt := fmt.Sprintf("clean command will clean nodes [%s], continue clean (y/n)?", strings.Join(deleteNodes, ","))
			result := Confirm(prompt)
			if !result {
				logger.Debug("clean nodes command is skip")
				goto all
			}
		}
		//只清除nodes
		i.Nodes = deleteNodes
	}
	//3. 删除所有节点
all:
	if len(deleteNodes) == 0 && len(deleteMasters) == 0 && CleanAll {
		if !CleanForce { // flase
			result := Confirm(`clean command will clean all masters and nodes, continue clean (y/n)?`)
			if !result {
				logger.Debug("clean all node command is skip")
				goto end
			}
		}
		// 所有master节点
		i.Masters = masters
		// 所有node节点
		i.Nodes = nodes
		i.cleanAll = true
	}
end:
	if len(i.Masters) == 0 && len(i.Nodes) == 0 {
		logger.Warn("clean nodes and masters is empty,please check your args and config.yaml.")
		os.Exit(-1)
	}
	i.CheckValid()
	i.Clean()
	if i.cleanAll {
		logger.Info("if clean all and clean kubeprince config")
		home, _ := os.UserHomeDir()
		cfgPath := home + defaultConfigPath
		sshcmd.Cmd("/bin/sh", "-c", "rm -rf "+cfgPath)
	}

}


//Clean clean cluster.
func (i *PrinceClean) Clean() {
	var wg sync.WaitGroup
	//s 是要删除的数据
	//全局是当前的数据
	if len(i.Nodes) > 0 {
		//1. 再删除nodes
		for _, node := range i.Nodes {
			wg.Add(1)
			go func(node string) {
				defer wg.Done()
				i.cleanNode(node)
			}(node)
		}
		wg.Wait()
	}
	if len(i.Masters) > 0 {
		//2. 先删除master
		for _, master := range i.Masters {
			wg.Add(1)
			go func(master string) {
				defer wg.Done()
				i.cleanMaster(master)
			}(master)
		}
		wg.Wait()
	}

}

func (i *PrinceClean) cleanNode(node string) {
	cleanRoute(node)
	clean(node)
	//remove node
	Nodes = SliceRemoveStr(Nodes, node)
	if !i.cleanAll {
		logger.Debug("clean node in master")
		if len(Masters) > 0 {
			hostname := isHostName(Masters[0], node)
			cmd := "kubectl delete node %s"
			_ = SSHConfig.CmdAsync(Masters[0], fmt.Sprintf(cmd, strings.TrimSpace(hostname)))
		}
	}
}

func (s *PrinceClean) cleanMaster(master string) {
	clean(master)
	//remove master
	Masters = SliceRemoveStr(Masters, master)
	if !s.cleanAll {
		logger.Debug("clean node in master")
		if len(Masters) > 0 {
			hostname := isHostName(Masters[0], master)
			cmd := "kubectl delete node %s"
			_ = SSHConfig.CmdAsync(Masters[0], fmt.Sprintf(cmd, strings.TrimSpace(hostname)))
		}
		//清空所有的nodes的数据
		yaml := ipvs.LvsStaticPodYaml(VIP, Masters, LvscareImage)
		var wg sync.WaitGroup
		for _, node := range Nodes {
			wg.Add(1)
			go func(node string) {
				defer wg.Done()
				_ = SSHConfig.CmdAsync(node, "rm -rf  /etc/kubernetes/manifests/kube-sealyun-lvscare*")
				_ = SSHConfig.CmdAsync(node, fmt.Sprintf("mkdir -p /etc/kubernetes/manifests && echo '%s' > /etc/kubernetes/manifests/kube-sealyun-lvscare.yaml", yaml))
			}(node)
		}
		wg.Wait()
	}
}

func clean(host string) {
	cmd := "kubeadm reset -f " + vlogToStr()
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = fmt.Sprintf(`sed -i '/kubectl/d;/kubeprince/d' /root/.bashrc`)
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = "modprobe -r ipip  && lsmod"
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = "rm -rf ~/.kube/ && rm -rf /etc/kubernetes/"
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = "rm -rf /etc/systemd/system/kubelet.service.d && rm -rf /etc/systemd/system/kubelet.service"
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = "rm -rf /usr/bin/kube* && rm -rf /usr/bin/crictl"
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = "rm -rf /etc/cni && rm -rf /opt/cni"
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = "rm -rf /var/lib/etcd && rm -rf /var/etcd"
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = fmt.Sprintf("sed -i \"/%s/d\" /etc/hosts ", ApiServer)
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = fmt.Sprint("rm -rf ~/kube")
	_ = SSHConfig.CmdAsync(host, cmd)
	//clean pki certs
	cmd = fmt.Sprint("rm -rf /etc/kubernetes/pki")
	_ = SSHConfig.CmdAsync(host, cmd)
	//clean kubeprince in /usr/bin/ except exec kubeprince
	cmd = fmt.Sprint("ps -ef |grep -v 'grep'|grep kubeprince >/dev/null || rm -rf /usr/bin/kubeprince")
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = fmt.Sprint("iptables -F &&  iptables -X &&  iptables -F -t nat &&  iptables -X -t nat")
	_ = SSHConfig.CmdAsync(host, cmd)
}

func cleanRoute(node string) {
	// clean route
	cmdRoute := fmt.Sprintf("kubeprince route --host %s", IpFormat(node))
	status := SSHConfig.CmdToString(node, cmdRoute, "")
	if status != "ok" {
		// 删除为 vip创建的路由。
		delRouteCmd := fmt.Sprintf("kubeprince route del --host %s --gateway %s", VIP, IpFormat(node))
		SSHConfig.CmdToString(node, delRouteCmd, "")
	}
}
