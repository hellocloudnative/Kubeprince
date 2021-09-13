package bin

import (
	 "Kubeprince/install"
	"github.com/spf13/cobra"
	"github.com/wonderivan/logger"
	"os"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Simplest way to update your kubernets HA cluster",
	Long:  `kubeprince clean --master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 --node 192.168.0.5 --user root --password your-server-password`,
	Example: exampleCleanCmd,
	Run: UpdateCmdFunc,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringSliceVar(&install.Masters, "master", []string{}, "kubernetes masters")
	updateCmd.Flags().StringSliceVar(&install.Nodes, "node", []string{}, "kubernetes nodes")
	updateCmd.PersistentFlags().BoolVarP(&install.UpdateForce, "force", "f", false, "if this is true, will no prompt")
	updateCmd.PersistentFlags().BoolVar(&install.UpdateAll, "all", false, "if this is true, update all ")
}

func UpdateCmdFunc(cmd *cobra.Command, args []string) {
	updateNodes := install.ParseIPs(install.Nodes)
	updateMasters := install.ParseIPs(install.Masters)
	c := &install.PrinceConfig{}
	err := c.Load(cfgFile)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
		//}
	}

	if ok, node := updateOrJoinNodeIsExistInCfgNodes(updateNodes, c.Masters); ok {
		logger.Error(`update master Use "kubeprince update--master %s" to clean it, exit...`, node)
		os.Exit(-1)
	}
	if ok, node := updateOrJoinNodeIsExistInCfgNodes(updateMasters, c.Nodes); ok {
		logger.Error(`update nodes Use "kubeprince update --node %s" to update it, exit...`, node)
		os.Exit(-1)
	}

	install.BuildUpdate(updateNodes, updateMasters)
	c.Dump(cfgFile)

}
// IsExistNodes
func updateOrJoinNodeIsExistInCfgNodes(updateOrJoinNodes []string, nodes []string) (bool, string) {
	for _, node := range nodes {
		for _, updateOrJoinNode := range updateOrJoinNodes{
			// 如果ips 相同. 则说明cfg配置文件已经存在该node.
			if node == updateOrJoinNode {
				return true, node
			}
		}
	}
	return false, ""
}


