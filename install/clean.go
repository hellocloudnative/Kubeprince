package install

func BuildClean(){
	hosts :=append(Masters,Nodes...)
	x := &PrinceInstaller{
		Hosts: hosts,
	}
	x.CheckCalid()
	x.Clean()


}

