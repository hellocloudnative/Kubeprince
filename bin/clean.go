package bin

import (
	"github.com/spf13/cobra"
	"Kubeprince/install"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Simplest way to clean your kubernets HA cluster",
	Long:  `K8sprince clean --master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 --node 192.168.0.5 --user root --password your-server-password`,
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildClean()
	},
}
func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringVar(&install.User, "user", "root", "servers user name for ssh")
	cleanCmd.Flags().StringVar(&install.Password, "password", "", "password for ssh")
	cleanCmd.Flags().StringVar(&install.PrivateKeyFile, "pk", "/root/.ssh/id_rsa", "private key for ssh")
	cleanCmd.Flags().StringVar(&install.ApiServer, "apiserver", "apiserver.cluster.local", "apiserver domain name")
	cleanCmd.Flags().StringSliceVar(&install.Masters, "master", []string{}, "kubernetes masters")
	cleanCmd.Flags().StringSliceVar(&install.Nodes, "node", []string{}, "kubernetes nodes")
}

