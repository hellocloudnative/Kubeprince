package bin

import (
	"Kubeprince/install"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init your kubernetes HA cluster",
	Long:  `K8sprince init --master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 --node 192.168.0.5 --user root --passwd your-password`,
	Run: func(cmd *cobra.Command, args []string) {
		 install.BuildInit()

	},
}

func init()  {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&install.User, "user", "root", "servers user name for ssh")
	initCmd.Flags().StringVar(&install.Password, "password", "", "password for ssh")
	initCmd.Flags().StringVar(&install.PrivateKeyFile, "pk", "/root/.ssh/id_rsa", "private key for ssh")
	initCmd.Flags().StringVar(&install.KubeadmFile, "kubeadm-config", "", "kubeadm-config.yaml template file")
	initCmd.Flags().StringVar(&install.ApiServer, "apiserver", "apiserver.cluster.local", "apiserver domain name")
	initCmd.Flags().StringVar(&install.VIP, "vip", "10.103.97.2", "virtual ip")
	initCmd.Flags().StringSliceVar(&install.Masters, "master", []string{}, "kubernetes masters")
	initCmd.Flags().StringSliceVar(&install.Nodes, "node", []string{}, "kubernetes nodes")
	initCmd.Flags().StringVar(&install.PkgUrl, "pkg-url", "", "download offline package url, or file localtion ex. /root/kube1.16.3.tar.gz")
	initCmd.Flags().StringVar(&install.Version, "version", "v1.16.3", "version is kubernetes version")
}
