package install


func BuildJoin() {
	i := &PrinceInstaller{
		Hosts: Nodes,
	}
	i.CheckCalid()
	i.SendPackage("kube")
	i.GeneratorToken()
	i.JoinNodes()
}
