package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/strix/maverik/pkg/maverik"
	"github.com/strix/maverik/pkg/questions"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "maverik",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if viper.Get("username") == nil && viper.Get("password") == nil {
			answers := questions.Ask()
			save := questions.ShouldSave()
			if save {
				// ******* REMOVE WHEN https://github.com/spf13/viper/pull/450 is merged
				if viper.ConfigFileUsed() == "" {
					home, _ := homedir.Dir()
					_, err := os.Create(home + "/.maverik.yaml")
					if err != nil {
						panic(err)
					}
				}
				// ******* END REMOVE **************************************************
				viper.Set("username", answers.Username)
				viper.Set("password", answers.Password)
				// ******* UNCOMMENT WHEN https://github.com/spf13/viper/pull/450 is merged
				// err = viper.SafeWriteConfig()
				// ******* END UNCOMMNENT *************************************************
				err := viper.WriteConfig()
				if err != nil {
					panic(err)
				}
				maverik.Login(answers.Username, answers.Password)
			}
		} else {
			maverik.Login(viper.GetString("username"), viper.GetString("password"))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		maverik.PrintSummary()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.maverik.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".maverik" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".maverik")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
