package bin

import (
"github.com/spf13/cobra"
"Kubeprince/cert"
)

type Flag struct {
	AltNames     []string
	NodeName     string
	ServiceCIDR  string
	NodeIP       string
	DNSDomain    string
	CertPath     string
	CertEtcdPath string
}

var config *Flag

// certCmd represents the cert command
var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "generate certs",
	Long:  `you can specify expire time`,
	Run: func(cmd *cobra.Command, args []string) {
		cert.GenerateCert(config.CertPath, config.CertEtcdPath, config.AltNames, config.NodeIP, config.NodeName, config.ServiceCIDR, config.DNSDomain)
	},
}

func init() {
	config = &Flag{}
	rootCmd.AddCommand(certCmd)

	certCmd.Flags().StringSliceVar(&config.AltNames, "alt-names", []string{}, "like zhangpengxuan.com or 10.103.97.2")
	certCmd.Flags().StringVar(&config.NodeName, "node-name", "", "like master0")
	certCmd.Flags().StringVar(&config.ServiceCIDR, "service-cidr", "", "like 10.103.97.2/24")
	certCmd.Flags().StringVar(&config.NodeIP, "node-ip", "", "like 10.103.97.2")
	certCmd.Flags().StringVar(&config.DNSDomain, "dns-domain", "cluster.local", "cluster dns domain")
	certCmd.Flags().StringVar(&config.CertPath, "cert-path", "/etc/kubernetes/pki", "kubernetes cert file path")
	certCmd.Flags().StringVar(&config.CertEtcdPath, "cert-etcd-path", "/etc/kubernetes/pki/etcd", "kubernetes etcd cert file path")
}
