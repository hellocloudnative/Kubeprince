package install

func BuildInit()  {
	Masters = ParseIPs(Masters)
	Nodes = ParseIPs(Nodes)
	// 所有master节点
	masters := Masters
	// 所有node节点
	nodes := Nodes
	hosts :=append(masters,nodes...)
	x := &PrinceInstaller{
		Hosts: hosts,
		Masters:   masters,
		Nodes:     nodes,
		Network:   Network,
		ApiServer: ApiServer,
	}
	x.CheckValid()
	x.Print()
	x.Sendkubeprince()
	x.SendPackage()
	x.Print("Initialization of the installation package is completed !")
	x.KubeadmConfigInstall()
	x.Print("SendPackage","kubeprince Configuration load completed !")
	x.GeneratorCert()
	x.CreateKubeconfig()
	x.InstallMaster0()
	x.Print("Install","master deployment completed !")
	if len(Masters) > 1 {
		x.JoinMasters(x.Masters[1:])
		x.Print( "JoinMasters","Adding the master is completed !")
	}
	if len(Nodes) > 0 {
		x.JoinNodes()
		x.Print( "Adding the node is completed !")
	}
	x.PrintFinish()


}