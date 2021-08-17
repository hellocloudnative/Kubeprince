package bin

import (
	"github.com/spf13/cobra"
	"Kubeprince/install"
	"github.com/wonderivan/logger"

	"os"
)
var exampleCleanCmd = `
	# clean  master
	kubeprince clean --master 192.168.0.2 \
	--master 192.168.0.3
  
	# clean  node  use --force/-f will be not prompt 
	kubeprince clean --node 192.168.0.4 \
	--node 192.168.0.5 --force

	# clean master and node
	kubeprince clean --master 192.168.0.2-192.168.0.3 \
 	--node 192.168.0.4-192.168.0.5
	
	# clean your kubernets HA cluster and use --force/-f will be not prompt (danger)
	kubeprince clean --all -f
`
// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Simplest way to clean your kubernets HA cluster",
	Long:  `kubeprince clean --master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 --node 192.168.0.5 --user root --password your-server-password`,
	Example: exampleCleanCmd,
	Run: CleanCmdFunc,
}
func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringSliceVar(&install.Masters, "master", []string{}, "kubernetes masters")
	cleanCmd.Flags().StringSliceVar(&install.Nodes, "node", []string{}, "kubernetes nodes")
	cleanCmd.PersistentFlags().BoolVarP(&install.CleanForce, "force", "f", false, "if this is true, will no prompt")
	cleanCmd.PersistentFlags().BoolVar(&install.CleanAll, "all", false, "if this is true, delete all ")
	cleanCmd.Flags().IntVar(&install.Vlog, "vlog", 0, "kubeadm log level")
}


func CleanCmdFunc(cmd *cobra.Command, args []string) {
	deleteNodes := install.ParseIPs(install.Nodes)
	deleteMasters := install.ParseIPs(install.Masters)
	c := &install.PrinceConfig{}
	err := c.Load(cfgFile)
	if err != nil {
		// comment: if cfgFile is not exist; do not use sealos clean something.
		// its danger for sealos do clean nodes without `~/.sealos/config.yaml`
		//// 判断错误是否为配置文件不存在
		//if errors.Is(err, os.ErrNotExist) {
		//	_, err = fmt.Fprint(os.Stdout, "Please enter the password to connect to the node:\n")
		//	if err != nil {
		//		logger.Error("fmt.Fprint err", err)
		//		os.Exit(-1)
		//	}
		//	passwordTmp, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		//	if err != nil {
		//		logger.Error("read password err", err)
		//		os.Exit(-1)
		//	}
		//	install.SSHConfig.Password = string(passwordTmp)
		//} else {
		logger.Error(err)
		os.Exit(-1)
		//}
	}

	if ok, node := deleteOrJoinNodeIsExistInCfgNodes(deleteNodes, c.Masters); ok {
		logger.Error(`clean master Use "kubeprince clean --master %s" to clean it, exit...`, node)
		os.Exit(-1)
	}
	if ok, node := deleteOrJoinNodeIsExistInCfgNodes(deleteMasters, c.Nodes); ok {
		logger.Error(`clean nodes Use "kubeprince clean --node %s" to clean it, exit...`, node)
		os.Exit(-1)
	}

	install.BuildClean(deleteNodes, deleteMasters)
	c.Dump(cfgFile)

}



// IsExistNodes
func deleteOrJoinNodeIsExistInCfgNodes(deleteOrJoinNodes []string, nodes []string) (bool, string) {
	for _, node := range nodes {
		for _, deleteOrJoinNode := range deleteOrJoinNodes {
			// 如果ips 相同. 则说明cfg配置文件已经存在该node.
			if node == deleteOrJoinNode {
				return true, node
			}
		}
	}
	return false, ""
}
