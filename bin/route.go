package bin

import (
	"github.com/spf13/cobra"

	"Kubeprince/install"
)

var (
	host      string
	gatewayIp string
)

func NewRouteCmd() *cobra.Command {
	// routeCmd represents the route command
	var cmd = &cobra.Command{
		Use:   "route",
		Short: "set default route gateway",
		Run:   RouteCmdFunc,
	}
	// check route for host
	cmd.Flags().StringVar(&host, "host", "", "route host ip address for iFace")
	cmd.AddCommand(NewDelRouteCmd())
	cmd.AddCommand(NewAddRouteCmd())
	return cmd
}

func init() {
	rootCmd.AddCommand(NewRouteCmd())
}

func NewAddRouteCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add",
		Short: "set route host via gateway",
		Run:   RouteAddCmdFunc,
	}
	// manually to set host via gateway
	cmd.Flags().StringVar(&host, "host", "", "route host ,ex ip route add host via gateway")
	cmd.Flags().StringVar(&gatewayIp, "gateway", "", "route gateway ,ex ip route add host via gateway")
	return cmd
}

func NewDelRouteCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "del",
		Short: "del route host via gateway, like ip route del host via gateway",
		Run:   RouteDelCmdFunc,
	}
	// manually to set host via gateway
	cmd.Flags().StringVar(&host, "host", "", "route host ,ex ip route del host via gateway")
	cmd.Flags().StringVar(&gatewayIp, "gateway", "", "route gateway ,ex ip route del host via gateway")
	return cmd
}

func RouteCmdFunc(cmd *cobra.Command, args []string) {
	r := install.GetRouteFlag(host, gatewayIp)
	r.CheckRoute()
}

func RouteAddCmdFunc(cmd *cobra.Command, args []string) {
	r := install.GetRouteFlag(host, gatewayIp)
	r.SetRoute()
}

func RouteDelCmdFunc(cmd *cobra.Command, args []string) {
	r := install.GetRouteFlag(host, gatewayIp)
	r.DelRoute()
}
