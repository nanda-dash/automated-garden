package pkg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/calvinmclean/automated-garden/garden-app/pkg/influxdb"
	"github.com/calvinmclean/automated-garden/garden-app/pkg/mqtt"
	"github.com/rs/xid"
)

// PlantAction collects all the possible actions for a Plant into a single struct so these can easily be
// received as one request
type PlantAction struct {
	Water *WaterAction `json:"water"`
}

// Execute is responsible for performing the actual individual actions in this PlantAction.
// The actions are executed in a deliberate order to be most intuitive for a user that wants
// to perform multiple actions with one request
func (action *PlantAction) Execute(g *Garden, p *Plant, mqttClient mqtt.Client, influxdbClient influxdb.Client) error {
	if action.Water != nil {
		if err := action.Water.Execute(g, p, mqttClient, influxdbClient); err != nil {
			return err
		}
	}
	return nil
}

// WaterAction is an action for watering a Plant for the specified amount of time
type WaterAction struct {
	Duration       int  `json:"duration"`
	IgnoreMoisture bool `json:"ignore_moisture"`
}

// WaterMessage is the message being sent over MQTT to the embedded garden controller
type WaterMessage struct {
	Duration      int    `json:"duration"`
	PlantID       xid.ID `json:"id"`
	PlantPosition int    `json:"plant_position"`
}

// Execute sends the message over MQTT to the embedded garden controller. Before doing this, it
// will first check if watering is set to skip and if the moisture value is below the threshold
// if configured
func (action *WaterAction) Execute(g *Garden, p *Plant, mqttClient mqtt.Client, influxdbClient influxdb.Client) error {
	if p.WateringStrategy.MinimumMoisture > 0 && !action.IgnoreMoisture {
		ctx, cancel := context.WithTimeout(context.Background(), influxdb.QueryTimeout)
		defer cancel()

		defer influxdbClient.Close()
		moisture, err := influxdbClient.GetMoisture(ctx, *p.PlantPosition, g.Name)
		if err != nil {
			return fmt.Errorf("error getting Plant's moisture data: %v", err)
		}
		if moisture > float64(p.WateringStrategy.MinimumMoisture) {
			return fmt.Errorf("moisture value %.2f%% is above threshold %d%%", moisture, p.WateringStrategy.MinimumMoisture)
		}
	}
	if p.SkipCount != nil && *p.SkipCount > 0 {
		*p.SkipCount--
		return nil
	}

	msg, err := json.Marshal(WaterMessage{
		Duration:      action.Duration,
		PlantID:       p.ID,
		PlantPosition: *p.PlantPosition,
	})
	if err != nil {
		return fmt.Errorf("unable to marshal WaterMessage to JSON: %v", err)
	}

	topic, err := mqttClient.WateringTopic(g.Name)
	if err != nil {
		return fmt.Errorf("unable to fill MQTT topic template: %v", err)
	}

	return mqttClient.Publish(topic, msg)
}

// GardenAction collects all the possible actions for a Garden into a single struct so these can easily be
// received as one request
type GardenAction struct {
	Light *LightAction `json:"light"`
	Stop  *StopAction  `json:"stop"`
}

// Execute is responsible for performing the actual individual actions in this GardenAction.
// The actions are executed in a deliberate order to be most intuitive for a user that wants
// to perform multiple actions with one request
func (action *GardenAction) Execute(g *Garden, mqttClient mqtt.Client) error {
	if action.Stop != nil {
		if err := action.Stop.Execute(g, mqttClient); err != nil {
			return err
		}
	}
	if action.Light != nil {
		if err := action.Light.Execute(g, mqttClient); err != nil {
			return err
		}
	}
	return nil
}

// LightAction is an action for turning on or off a light for the Garden. The State field is optional and it will just toggle
// the current state if left empty.
type LightAction struct {
	State string `json:"state"`
}

// Execute sends an MQTT message to the garden controller to change the state of the light
func (action *LightAction) Execute(g *Garden, mqttClient mqtt.Client) error {
	msg, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("unable to marshal LightAction to JSON: %v", err)
	}

	topic, err := mqttClient.LightTopic(g.Name)
	if err != nil {
		return fmt.Errorf("unable to fill MQTT topic template: %v", err)
	}

	return mqttClient.Publish(topic, msg)
}

// StopAction is an action for stopping watering of a Plant. It doesn't stop watering a specific Plant, only what is
// currently watering and optionally clearing the queue of Plants to water.
type StopAction struct {
	All bool `json:"all"`
}

// Execute sends the message over MQTT to the embedded garden controller
func (action *StopAction) Execute(g *Garden, mqttClient mqtt.Client) error {
	topicFunc := mqttClient.StopTopic
	if action.All {
		topicFunc = mqttClient.StopAllTopic
	}
	topic, err := topicFunc(g.Name)
	if err != nil {
		return fmt.Errorf("unable to fill MQTT topic template: %v", err)
	}

	return mqttClient.Publish(topic, []byte("no message"))
}
