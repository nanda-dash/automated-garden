package cmd

import (
	"time"

	"github.com/calvinmclean/automated-garden/garden-app/controller"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	topicPrefix       string
	numZones          int
	moistureStrategy  string
	moistureValue     int
	moistureInterval  time.Duration
	publishWaterEvent bool
	publishHealth     bool
	healthInterval    time.Duration
	enableUI          bool

	controllerCommand = &cobra.Command{
		Use:   "controller",
		Short: "Run a mock garden-controller",
		Long:  `Subscribes on a MQTT topic to act as a mock garden-controller for testing purposes`,
		Run:   Controller,
	}
)

func init() {
	controllerCommand.Flags().StringVarP(&topicPrefix, "topic", "t", "test-garden", "MQTT topic prefix of the garden-controller")
	viper.BindPFlag("topic_prefix", controllerCommand.Flags().Lookup("topic"))

	controllerCommand.Flags().IntVarP(&numZones, "zones", "z", 0, "Number of Zones for which moisture data should be emulated")
	viper.BindPFlag("num_zones", controllerCommand.Flags().Lookup("zones"))

	controllerCommand.Flags().StringVar(&moistureStrategy, "moisture-strategy", "random", "Strategy for creating moisture data")
	controllerCommand.RegisterFlagCompletionFunc("moisture-strategy", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"random", "constant", "increasing", "decreasing"}, cobra.ShellCompDirectiveDefault
	})
	viper.BindPFlag("moisture_strategy", controllerCommand.Flags().Lookup("moisture-strategy"))

	controllerCommand.Flags().IntVar(&moistureValue, "moisture-value", 100, "The value, or starting value, to use for moisture data publishing")
	viper.BindPFlag("moisture_value", controllerCommand.Flags().Lookup("moisture-value"))

	controllerCommand.Flags().DurationVar(&moistureInterval, "moisture-interval", 10*time.Second, "Interval between moisture data publishing")
	viper.BindPFlag("moisture_interval", controllerCommand.Flags().Lookup("moisture-interval"))

	controllerCommand.Flags().BoolVar(&publishWaterEvent, "publish-water-event", true, "Whether or not watering events should be published for logging")
	viper.BindPFlag("publish_water_event", controllerCommand.Flags().Lookup("publish-water-event"))

	controllerCommand.Flags().BoolVar(&publishHealth, "publish-health", true, "Whether or not to publish health data every minute")
	viper.BindPFlag("publish_health", controllerCommand.Flags().Lookup("publish-health"))

	controllerCommand.Flags().DurationVar(&healthInterval, "health-interval", time.Minute, "Interval between health data publishing")
	viper.BindPFlag("health_interval", controllerCommand.Flags().Lookup("health-interval"))

	controllerCommand.Flags().BoolVar(&enableUI, "enable-ui", true, "Enable tview UI for nicer output")
	viper.BindPFlag("enable_ui", controllerCommand.Flags().Lookup("enable-ui"))

	rootCommand.AddCommand(controllerCommand)
}

// Controller will start up the mock garden-controller
func Controller(cmd *cobra.Command, args []string) {
	var config controller.Config
	if err := viper.Unmarshal(&config); err != nil {
		cmd.PrintErrln("unable to read config from file: ", err)
		return
	}
	config.LogLevel = parsedLogLevel

	controller.Start(config)
}
