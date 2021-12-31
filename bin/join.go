package bin

import (
	"github.com/spf13/cobra"
	"Kubeprince/install"
	"github.com/wonderivan/logger"
	"os"
)

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Simplest way to join your kubernets HA cluster",
	Long:  `kubeprince join --master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 --node 192.168.0.5 `,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(install.Masters) == 0 && len(install.Nodes) == 0 {
			logger.Error("this command is join feature,master and node is empty at the same time.please check your args in command.")
			cmd.Help()
			os.Exit(0)
		}
	},
	Run: BuildJoin,
}

func init() {
	rootCmd.AddCommand(joinCmd)
	joinCmd.Flags().StringSliceVar(&install.Masters, "master", []string{}, "kubernetes multi-master ex. 192.168.0.5-192.168.0.5")
	joinCmd.Flags().StringSliceVar(&install.Nodes, "node", []string{}, "kubernetes multi-nodes ex. 192.168.0.5-192.168.0.5")
	joinCmd.Flags().StringVar(&install.Containers, "containers", "isulad", "isulad, docker..")
	joinCmd.Flags().IntVar(&install.Vlog, "vlog", 0, "kubeadm log level")
}

func BuildJoin (cmd *cobra.Command, args []string) {
	beforeNodes := install.ParseIPs(install.Nodes)
	beforeMasters := install.ParseIPs(install.Masters)

	c := &install.PrinceConfig{}
	err := c.Load(cfgFile)
	if err != nil {
		logger.Error(err)
		c.ShowDefaultConfig()
		os.Exit(0)
	}

	cfgNodes := append(c.Masters, c.Nodes...)
	joinNodes := append(beforeNodes, beforeMasters...)

	if ok, node := deleteOrJoinNodeIsExistInCfgNodes(joinNodes, cfgNodes); ok {
		logger.Error(`[%s] has already exist in your cluster. please check.`, node)
		os.Exit(-1)
	}

	install.BuildJoin(beforeMasters, beforeNodes)
	c.Dump(cfgFile)
}
