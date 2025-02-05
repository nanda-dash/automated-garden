package controller

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
)

const (
	configTemplate = `#ifndef config_h
#define config_h

#define TOPIC_PREFIX "{{ .TopicPrefix }}"

#define QUEUE_SIZE 10

#define ENABLE_WIFI
#ifdef ENABLE_WIFI
#define MQTT_ADDRESS "{{ .MQTTConfig.Broker }}"
#define MQTT_PORT {{ .MQTTConfig.Port }}
#define MQTT_CLIENT_NAME TOPIC_PREFIX
#define MQTT_WATER_TOPIC TOPIC_PREFIX"/command/water"
#define MQTT_STOP_TOPIC TOPIC_PREFIX"/command/stop"
#define MQTT_STOP_ALL_TOPIC TOPIC_PREFIX"/command/stop_all"
#define MQTT_LIGHT_TOPIC TOPIC_PREFIX"/command/light"
#define MQTT_LIGHT_DATA_TOPIC TOPIC_PREFIX"/data/light"
#define MQTT_WATER_DATA_TOPIC TOPIC_PREFIX"/data/water"

{{ if .PublishHealth }}
#define ENABLE_MQTT_HEALTH
#ifdef ENABLE_MQTT_HEALTH
#define MQTT_HEALTH_DATA_TOPIC TOPIC_PREFIX"/data/health"
#define HEALTH_PUBLISH_INTERVAL {{ milliseconds .HealthInterval }}
#endif
{{ end }}

#define ENABLE_MQTT_LOGGING
#ifdef ENABLE_MQTT_LOGGING
#define MQTT_LOGGING_TOPIC TOPIC_PREFIX"/data/logs"
#endif

#define JSON_CAPACITY 48
#endif

{{ if .DisableWatering }}
#define DISABLE_WATERING
{{ end -}}
#define NUM_ZONES {{ len .Zones }}
#define ZONES { {{ range $index, $z := .Zones }}{{if $index}}, {{end}}{ {{ $z.PumpPin }}, {{ $z.ValvePin }}, {{ or $z.ButtonPin "GPIO_NUM_MAX" }}, {{ or $z.MoistureSensorPin "GPIO_NUM_MAX" }} }{{ end }} }
#define DEFAULT_WATER_TIME {{ milliseconds  .DefaultWaterTime }}

{{ if .LightPin }}
#define LIGHT_PIN {{ .LightPin }}
{{ end }}

{{ if .EnableButtons }}
#define ENABLE_BUTTONS
#ifdef ENABLE_BUTTONS
#define STOP_BUTTON_PIN {{ .StopButtonPin }}
#endif
{{ end }}

{{ if .EnableMoistureSensor }}
#ifdef ENABLE_MOISTURE_SENSORS AND ENABLE_WIFI
#define MQTT_MOISTURE_DATA_TOPIC TOPIC_PREFIX"/data/moisture"
#define MOISTURE_SENSOR_AIR_VALUE 3415
#define MOISTURE_SENSOR_WATER_VALUE 1362
#define MOISTURE_SENSOR_INTERVAL {{ milliseconds .MoistureInterval }}
#endif
{{ end -}}

{{ if .PublishTemperatureHumidity }}
#define ENABLE_DHT22
#ifdef ENABLE_DHT22
#define MQTT_TEMPERATURE_DATA_TOPIC TOPIC_PREFIX"/data/temperature"
#define MQTT_HUMIDITY_DATA_TOPIC TOPIC_PREFIX"/data/humidity"
#define DHT22_PIN {{ .TemperatureHumidityPin }}
#define DHT22_INTERVAL {{ milliseconds .TemperatureHumidityInterval }}
#endif
{{ end -}}
#endif
`
	wifiConfigTemplate = `#ifndef wifi_config_h
#define wifi_config_h

#define SSID "{{ .SSID }}"
#define PASSWORD "{{ .Password }}"

#endif
`
)

// WifiConfig holds WiFi connection details
type WifiConfig struct {
	SSID     string `mapstructure:"ssid"`
	Password string `mapstructure:"password"`
}

// ZoneConfig has the configuration details for controlling hardware pins
type ZoneConfig struct {
	PumpPin           string `mapstructure:"pump_pin" survey:"pump_pin"`
	ValvePin          string `mapstructure:"valve_pin" survey:"valve_pin"`
	ButtonPin         string `mapstructure:"button_pin" survey:"button_pin"`
	MoistureSensorPin string `mapstructure:"moisture_sensor_pin" survey:"moisture_sensor_pin"`
}

// GenerateConfig will create config.h and wifi_config.h based on the provided configurations. It can optionally write to files
// instead of stdout
func GenerateConfig(config Config, writeFile, genWifiConfig, genMainConfig, overwrite, interactive bool) {
	logger := setupLogger(config.LogConfig)

	if interactive {
		err := survey.AskOne(&survey.Confirm{
			Message: "Generate 'config.h'?",
			Default: genMainConfig,
		}, &genMainConfig)
		if err != nil {
			logger.WithError(err).Error("survey error")
			return
		}
	}

	if genMainConfig {
		logger.Debug("generating 'config.h'")
		mainConfig, err := generateMainConfig(config, interactive)
		if err != nil {
			logger.WithError(err).Error("error generating 'config.h'")
			return
		}
		err = writeOutput(logger, mainConfig, "config.h", writeFile, overwrite, interactive)
		if err != nil {
			logger.WithError(err).Error("error generating 'config.h'")
			return
		}
	}

	if interactive {
		err := survey.AskOne(&survey.Confirm{
			Message: "Generate 'wifi_config.h'?",
			Default: genWifiConfig,
		}, &genWifiConfig)
		if err != nil {
			logger.WithError(err).Error("survey error")
			return
		}
	}

	if genWifiConfig {
		logger.Debug("generating 'wifi_config.h'")
		wifiConfig, err := generateWiFiConfig(config.WifiConfig, interactive)
		if err != nil {
			logger.WithError(err).Error("error generating 'wifi_config.h'")
			return
		}
		err = writeOutput(logger, wifiConfig, "wifi_config.h", writeFile, overwrite, interactive)
		if err != nil {
			logger.WithError(err).Error("error generating 'wifi_config.h'")
			return
		}
	}
}

func writeOutput(logger *logrus.Logger, content, filename string, writeFile, overwrite, interactive bool) error {
	logger.WithFields(logrus.Fields{
		"filename":       filename,
		"write_file":     writeFile,
		"overwrite_file": overwrite,
	}).Debug("writing output to file")

	if interactive {
		err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Write generated config to %q?", filename),
			Default: writeFile,
		}, &writeFile)
		if err != nil {
			return err
		}
	}

	file := os.Stdout
	// if configured to write to a file, replace os.Stdout with file
	if writeFile {
		// if overwrite is false, first check if file exists and error if it does
		if !overwrite {
			_, err := os.Stat(filename)
			if err == nil {
				if interactive {
					err := survey.AskOne(&survey.Confirm{
						Message: fmt.Sprintf("Overwrite existing %q?", filename),
						Default: overwrite,
					}, &overwrite)
					if err != nil {
						return err
					}
				}
				if !overwrite {
					return fmt.Errorf("file %q exists, use --force to overwrite", filename)
				}
			}
		}

		var err error
		file, err = os.Create(filename)
		if err != nil {
			return err
		}
	}

	_, err := file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func generateMainConfig(config Config, interactive bool) (string, error) {
	if interactive {
		err := configPrompts(&config)
		if err != nil {
			return "", err
		}
	}

	milliseconds := func(interval time.Duration) string {
		return fmt.Sprintf("%d", interval.Milliseconds())
	}
	t := template.Must(template.
		New("config.h").
		Funcs(template.FuncMap{"milliseconds": milliseconds}).
		Parse(configTemplate))

	var result bytes.Buffer
	data := config
	err := t.Execute(&result, data)
	if err != nil {
		return "", err
	}
	return removeExtraNewlines(result.String()), nil
}

func generateWiFiConfig(config WifiConfig, interactive bool) (string, error) {
	qs := []*survey.Question{
		{
			Name: "ssid",
			Prompt: &survey.Input{
				Message: "WiFi SSID",
				Default: config.SSID,
				Help:    "this is the name of your WiFi network",
			},
			Validate: survey.Required,
		},
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Password",
				Help:    "this is your WiFi password",
			},
			Validate: survey.Required,
		},
	}

	// if not interactive, but password is missing, turn interactive with password question only
	if config.Password == "" && !interactive {
		qs = qs[1:]
		interactive = true
	}

	if interactive {
		err := survey.Ask(qs, &config)
		if err != nil {
			return "", fmt.Errorf("error in survey response: %w", err)
		}
	}

	t := template.Must(template.New("wifi_config.h").Parse(wifiConfigTemplate))
	var result bytes.Buffer
	err := t.Execute(&result, config)
	if err != nil {
		return "", err
	}
	return removeExtraNewlines(result.String()), nil
}

func removeExtraNewlines(input string) string {
	return regexp.MustCompile(`(?m)^\n{2,}`).ReplaceAllLiteralString(input, "\n")
}
