package bin

import (
"github.com/spf13/cobra"

"Kubeprince/install"
)

// ipvsCmd represents the ipvs command
var ipvsCmd = &cobra.Command{
	Use:   "ipvs",
	Short: "kubeprince create or care local ipvs lb",
	Run: func(cmd *cobra.Command, args []string) {
		install.Ipvs.VsAndRsCare()
	},
}

var clean bool

func init() {
	rootCmd.AddCommand(ipvsCmd)

	// Here you will define your flags and configuration settings.
	ipvsCmd.Flags().BoolVar(&install.Ipvs.RunOnce, "run-once", false, "is run once mode")
	ipvsCmd.Flags().BoolVarP(&install.Ipvs.Clean, "clean", "c", true, " clean Vip ipvs rule before join node, if Vip has no ipvs rule do nothing.")
	ipvsCmd.Flags().StringVar(&install.Ipvs.VirtualServer, "vs", "", "virturl server like 10.54.0.2:6443")
	ipvsCmd.Flags().StringSliceVar(&install.Ipvs.RealServer, "rs", []string{}, "virturl server like 192.168.0.2:6443")

	ipvsCmd.Flags().StringVar(&install.Ipvs.HealthPath, "health-path", "/healthz", "health check path")
	ipvsCmd.Flags().StringVar(&install.Ipvs.HealthSchem, "health-schem", "https", "health check schem")
	ipvsCmd.Flags().Int32Var(&install.Ipvs.Interval, "interval", 5, "health check interval, unit is sec.")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ipvsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ipvsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
