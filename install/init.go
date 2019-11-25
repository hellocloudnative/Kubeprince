package install

func BuildInit()  {
	hosts :=append(Masters,Nodes...)
	x := &PrinceInstaller{
		Hosts: hosts,
	}
	x.CheckCalid()
	x.Print()
	x.SendPackage("kube")
	x.Print("SendPackage")
	x.KubeadmConfigInstall()
	x.Print("K8Sprinceconfiginstall")
	x.InstallMaster0()
	x.Print("InstallMaster0")
	if len(Masters) > 1 {
		x.JoinMasters()
		x.Print("SendPackage", "KubeadmConfigInstall", "InstallMaster0", "JoinMasters")
	}
	if len(Nodes) > 0 {
		x.JoinNodes()
		x.Print("SendPackage", "KubeadmConfigInstall", "InstallMaster0", "JoinMasters", "JoinNodes")
	}
	x.PrintFinish()


}