package cmd

import (
	"github.com/spf13/cobra"
	"github.com/strix/maverik/pkg/maverik"
)

// transactionsCmd represents the summary command
var transactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Print recent transactions",
	Long:  `Prints the recent transaction (currently 60 days) and point accrual associated with them.`,
	Run: func(cmd *cobra.Command, args []string) {
		maverik.PrintTransactions()
	},
}

func init() {
	rootCmd.AddCommand(transactionsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transactionsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transactionsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
