
package cmd

import (
		"github.com/spf13/cobra"
		"../ping"
)

var waitTimeMs int

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping a hostname or an IP address (only IPV4)",
		Long: "You can ping a hostname like google.com, or an IPV4 address like 1.1.1.1",
		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
				ping.StartPinging(args[0], waitTimeMs)
		},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	pingCmd.Flags().IntVarP(&waitTimeMs, "wait", "w", 1000, "Number of milliseconds to wait between sending pings")
}
