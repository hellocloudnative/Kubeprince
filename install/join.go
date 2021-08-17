package install

import (
	"Kubeprince/cert"
	"Kubeprince/ipvs"
	"sync"
)

//func BuildJoin() {
//	i := &PrinceInstaller{
//		Hosts: Nodes,
//	}
//	i.CheckCalid()
//	i.JoinNodes()
//}
//BuildJoin is
func BuildJoin(joinMasters, joinNodes []string) {
	if len(joinMasters) > 0 {
		joinMastersFunc(joinMasters)
	}
	if len(joinNodes) > 0 {
		joinNodesFunc(joinNodes)
	}
}


func joinMastersFunc(joinMasters []string) {
	masters := Masters
	nodes := Nodes
	x := &PrinceInstaller{
		Hosts:     joinMasters,
		Masters:   masters,
		Nodes:     nodes,
		Network:   Network,
		ApiServer: ApiServer,
	}
	x.CheckValid()
	x.Sendkubeprince()
	x.SendPackage()
	x.GeneratorCert()
	x.JoinMasters(joinMasters)
	//master join to MasterIPs
	Masters = append(Masters, joinMasters...)
	x.lvscare()

}

//joinNodesFunc is join nodes func
func joinNodesFunc(joinNodes []string) {
	// 所有node节点
	nodes := joinNodes
	x := &PrinceInstaller{
		Hosts:   nodes,
		Masters: Masters,
		Nodes:   nodes,
	}
	x.CheckValid()
	x.Sendkubeprince()
	x.SendPackage()
	x.GeneratorCert()
	x.JoinNodes()
	//node join to NodeIPs
	Nodes = append(Nodes, joinNodes...)
}

func (s *PrinceInstaller) sendKubeConfigFile(hosts []string, kubeFile string) {
	absKubeFile := cert.KubernetesDir + "/" + kubeFile
	sealosKubeFile := cert.KubeprinceConfigDir + "/" + kubeFile
	var wg sync.WaitGroup
	for _, node := range hosts {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			SSHConfig.CopyLocalToRemote(node, sealosKubeFile, absKubeFile)
		}(node)
	}
	wg.Wait()
}

func (s *PrinceInstaller) lvscare() {
	var wg sync.WaitGroup
	for _, node := range s.Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			yaml := ipvs.LvsStaticPodYaml(VIP, Masters, LvscareImage)
			_ = SSHConfig.Cmd(node, "rm -rf  /etc/kubernetes/manifests/kube-sealyun-lvscare* || :")
			SSHConfig.CopyConfigFile(node, "/etc/kubernetes/manifests/kube-sealyun-lvscare.yaml", []byte(yaml))
		}(node)
	}

	wg.Wait()
}
