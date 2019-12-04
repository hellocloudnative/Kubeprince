package bin

import (
	"github.com/spf13/cobra"
	"Kubeprince/install"
)

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Simplest way to join your kubernets HA cluster",
	Long:  `kubeprince join --master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 --node 192.168.0.5 --vip 192.168.0.1  --user root --passwd your-server-password --pkg-url /root/kube1.14.1.tar.gz`,
	Run: func(cmd *cobra.Command, args []string) {
		install.BuildJoin()
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)
	joinCmd.Flags().StringVar(&install.User, "user", "root", "servers user name for ssh")
	joinCmd.Flags().StringVar(&install.Password, "password", "", "password for ssh")
	joinCmd.Flags().StringVar(&install.PrivateKeyFile, "pk", "/root/.ssh/id_rsa", "private key for ssh")

	joinCmd.Flags().StringVar(&install.ApiServer, "apiserver", "apiserver.cluster.local", "apiserver domain name")
	joinCmd.Flags().StringVar(&install.VIP, "vip", "10.103.97.2", "virtual ip")
	joinCmd.Flags().StringSliceVar(&install.Masters, "master", []string{}, "kubernetes masters")
	joinCmd.Flags().StringSliceVar(&install.Nodes, "node", []string{}, "kubernetes nodes")

	joinCmd.Flags().StringVar(&install.PkgUrl, "pkg-url", "", "download offline pakage url, or file localtion ex.")
}
