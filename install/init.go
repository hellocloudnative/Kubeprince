package install

func BuildInit()  {
	hosts :=append(Masters,Nodes...)
	x := &PrinceInstaller{
		Hosts: hosts,
	}
	x.CheckCalid()
	x.Print()
	x.SendPackage("kube")
	x.Print("Initialization of the installation package is completed !")
	x.KubeadmConfigInstall()
	x.Print("kubeprince Configuration load completed !")
	x.InstallMaster0()
	x.Print("master deployment completed !")
	if len(Masters) > 1 {
		x.JoinMasters()
		x.Print( "Adding the master is completed !")
	}
	if len(Nodes) > 0 {
		x.JoinNodes()
		x.Print( "Adding the node is completed !")
	}
	x.PrintFinish()


}