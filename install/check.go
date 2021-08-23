package install

import (
	"github.com/wonderivan/logger"
	"os"
	"fmt"
)
// SetHosts set hosts. if can't access to hostName, set /etc/hosts
func SetHosts(hostIP, hostName string) {
	cmd := fmt.Sprintf("cat /etc/hosts |grep %s || echo '%s %s' >> /etc/hosts", hostName, IpFormat(hostIP), hostName)
	SSHConfig.CmdAsync(hostIP, cmd)
}

func (p  *PrinceInstaller) CheckValid() {
	hosts := append(p.Masters,p.Nodes...)
	if len(p.Hosts) == 0 && len(hosts) == 0 {
		p.Print("Fail")
		logger.Error("hosts not allow empty")
		os.Exit(1)
	}
	if SSHConfig.User == "" {
		p.Print("Fail")
		logger.Error("user not allow empty")
		os.Exit(1)
	}
	dict := make(map[string]bool)
	var errList []string
	for _,h :=range p.Hosts{
		//获取主机名
		hostname := SSHConfig.CmdToString(h, "hostname", "")
		if hostname == "" {
			logger.Error("[%s] ------------ check error", h)
			os.Exit(1)
		} else {
			SetHosts(h, hostname)
			if _, ok := dict[hostname]; !ok {
				//不冲突, 主机名加入字典
				dict[hostname] = true
			} else {
				logger.Error("duplicate hostnames is not allowed")
				os.Exit(1)
			}
			logger.Info("[%s]  ------------ check ok", h)
		}
		if p.Network == "cilium" {
			if err := SSHConfig.CmdAsync(h, "uname -r | grep 5 | awk -F. '{if($2>3)print \"ok\"}' | grep ok && exit 0 || exit 1"); err != nil {
				logger.Error("[%s] ------------ check kernel version  < 5.3", h)
				os.Exit(1)
			}
			if err := SSHConfig.CmdAsync(h, "mount bpffs -t bpf /sys/fs/bpf && mount | grep /sys/fs/bpf && exit 0 || exit 1"); err != nil {
				logger.Error("[%s] ------------ mount  bpffs err", h)
				os.Exit(1)
			}
		}
		dockerExist := SSHConfig.CmdToString(h, "command -v dockerd &> /dev/null && echo yes || :", "")
		if dockerExist == "yes" {
			errList = append(errList, h)
		}
		containerdExist := SSHConfig.CmdToString(h, "command -v containerd &> /dev/null && echo yes || :", "")
		if containerdExist == "yes" {
			errList = append(errList, h)
		}
		podmanExist := SSHConfig.CmdToString(h, "command -v podmanExist &> /dev/null && echo yes || :", "")
		if podmanExist == "yes" {
			errList = append(errList, h)
		}
		if len(errList) >= 1 {
			os.Exit(-1)
		}
	}
}