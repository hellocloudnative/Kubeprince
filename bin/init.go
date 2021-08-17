package bin

import (
	"Kubeprince/install"
	"github.com/spf13/cobra"
	"Kubeprince/pkg/logger"
	"os"
)
var exampleInit = `
	# init with password with three master one node
	kubeprince init --passwd your-server-password  \
	--master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 \
	--node 192.168.0.5 --user root \
	--version v1.18.0 --pkg-url=/root/kube1.18.5.tar.gz 
	
	# init with pk-file , when your server have different password
	kubeprince init --pk /root/.ssh/id_rsa \
	--master 192.168.0.2 --node 192.168.0.5 --user root \
	--version v1.18.0 --pkg-url=/root/kube1.18.5.tar.gz 

	# when use multi network. set a can-reach with --interface 
 	kubeprince init --interface 192.168.0.254 \
	--master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 \
	--node 192.168.0.5 --user root --passwd your-server-password \
	--version v1.18.0 --pkg-url=/root/kube1.18.5.tar.gz 
	
	# when your interface is not "eth*|en*|em*" like.
	kubeprince init --interface your-interface-name \
	--master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 \
	--node 192.168.0.5 --user root --passwd your-server-password \
	--version v1.18.0 --pkg-url=/root/kube1.18.5.tar.gz 
`
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init your kubernetes HA cluster",
	Long:  `kubeprince init --master 192.168.0.2 --master 192.168.0.3 --master 192.168.0.4 --node 192.168.0.5 --user root --passwd your-password`,
	Example: exampleInit,
	Run: func(cmd *cobra.Command, args []string) {
		c := &install.PrinceConfig{}
		// 没有重大错误可以直接保存配置. 但是apiservercertsans为空. 但是不影响用户 clean
		// 如果用户指定了配置文件,并不使用--master, 这里就不dump, 需要使用load获取配置文件了.
		if cfgFile != "" && len(install.Masters) == 0 {
			err := c.Load(cfgFile)
			if err != nil {
				logger.Error("load cfgFile %s err: %q", cfgFile, err)
				os.Exit(1)
			}
		} else {
			c.Dump(cfgFile)
		}
		install.BuildInit()
		// 安装完成后生成完整版
		c.Dump(cfgFile)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		// 使用了cfgFile 就不进行preRun了
		if cfgFile == "" && install.ExitInitCase() {
			cmd.Help()
			os.Exit(install.ErrorExitOSCase)
		}
	},
}


func init()  {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&install.SSHConfig.User, "user", "root", "servers user name for ssh")
	initCmd.Flags().StringVar(&install.SSHConfig.Password, "password", "", "password for ssh")
	initCmd.Flags().StringVar(&install.SSHConfig.PkFile, "pk", "/root/.ssh/id_rsa", "private key for ssh")
	initCmd.Flags().StringVar(&install.SSHConfig.PkPassword, "pk-passwd", "", "private key password for ssh")
	initCmd.Flags().StringVar(&install.KubeadmFile, "kubeadm-config", "", "kubeadm-config.yaml template file")
	initCmd.Flags().StringVar(&install.ApiServer, "apiserver", "apiserver.cluster.local", "apiserver domain name")
	initCmd.Flags().StringVar(&install.VIP, "vip", "10.103.97.2", "virtual ip")
	initCmd.Flags().StringVar(&install.Repo, "repo", "k8s.gcr.io", "choose a container registry to pull control plane images from")
	initCmd.Flags().StringVar(&install.PodCIDR, "podcidr", "100.64.0.0/10", "Specify range of IP addresses for the pod network")
	initCmd.Flags().StringVar(&install.SvcCIDR, "svccidr", "10.96.0.0/12", "Use alternative range of IP address for service VIPs")
	initCmd.Flags().StringSliceVar(&install.Masters, "master", []string{}, "kubernetes masters")
	initCmd.Flags().StringSliceVar(&install.Nodes, "node", []string{}, "kubernetes nodes")
	initCmd.Flags().StringVar(&install.PkgUrl, "pkg-url", "", "download offline package url, or file localtion ex. /root/kube1.18.5.tar.gz")
	initCmd.Flags().StringVar(&install.Version, "version", "v1.18.5", "version is kubernetes version")
}
