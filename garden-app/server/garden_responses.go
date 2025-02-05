package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/calvinmclean/automated-garden/garden-app/pkg"
)

// GardenResponse is used to represent a Garden in the response body with the additional Moisture data
// and hypermedia Links fields
type GardenResponse struct {
	*pkg.Garden
	NextLightAction         *NextLightAction         `json:"next_light_action,omitempty"`
	Health                  *pkg.GardenHealth        `json:"health,omitempty"`
	TemperatureHumidityData *TemperatureHumidityData `json:"temperature_humidity_data,omitempty"`
	NumPlants               uint                     `json:"num_plants"`
	NumZones                uint                     `json:"num_zones"`
	Plants                  Link                     `json:"plants"`
	Zones                   Link                     `json:"zones"`
	Links                   []Link                   `json:"links,omitempty"`
}

// NextLightAction contains the time and state for the next scheduled LightAction
type NextLightAction struct {
	Time  *time.Time     `json:"time"`
	State pkg.LightState `json:"state"`
}

// TemperatureHumidityData has the temperature and humidity of the Garden
type TemperatureHumidityData struct {
	TemperatureCelsius float64 `json:"temperature_celsius"`
	HumidityPercentage float64 `json:"humidity_percentage"`
}

// NewGardenResponse creates a self-referencing GardenResponse
func (gr GardensResource) NewGardenResponse(ctx context.Context, garden *pkg.Garden, links ...Link) *GardenResponse {
	plantsPath := fmt.Sprintf("%s/%s%s", gardenBasePath, garden.ID, plantBasePath)
	zonesPath := fmt.Sprintf("%s/%s%s", gardenBasePath, garden.ID, zoneBasePath)
	response := &GardenResponse{
		Garden:    garden,
		NumPlants: garden.NumPlants(),
		NumZones:  garden.NumZones(),
		Plants:    Link{"collection", plantsPath},
		Zones:     Link{"collection", zonesPath},
	}
	response.Links = append(links,
		Link{
			"self",
			fmt.Sprintf("%s/%s", gardenBasePath, garden.ID),
		},
	)

	if garden.EndDated() {
		return response
	}

	response.Links = append(response.Links,
		Link{
			"plants",
			plantsPath,
		},
		Link{
			"zones",
			zonesPath,
		},
		Link{
			"action",
			fmt.Sprintf("%s/%s/action", gardenBasePath, garden.ID),
		},
	)

	response.Health = garden.Health(ctx, gr.influxdbClient)

	if garden.LightSchedule != nil {
		nextOnTime := gr.worker.GetNextLightTime(garden, pkg.LightStateOn)
		nextOffTime := gr.worker.GetNextLightTime(garden, pkg.LightStateOff)
		if nextOnTime != nil && nextOffTime != nil {
			// If the nextOnTime is before the nextOffTime, that means the next light action will be the ON action
			if nextOnTime.Before(*nextOffTime) {
				response.NextLightAction = &NextLightAction{
					Time:  nextOnTime,
					State: pkg.LightStateOn,
				}
			} else {
				response.NextLightAction = &NextLightAction{
					Time:  nextOffTime,
					State: pkg.LightStateOff,
				}
			}
		} else if nextOnTime != nil {
			response.NextLightAction = &NextLightAction{
				Time:  nextOnTime,
				State: pkg.LightStateOn,
			}
		} else if nextOffTime != nil {
			response.NextLightAction = &NextLightAction{
				Time:  nextOffTime,
				State: pkg.LightStateOff,
			}
		}
	}

	if garden.HasTemperatureHumiditySensor() {
		t, h, err := gr.influxdbClient.GetTemperatureAndHumidity(ctx, garden.TopicPrefix)
		if err != nil {
			logger := getLoggerFromContext(ctx).WithField(gardenIDLogField, garden.ID.String())
			logger.WithError(err).Error("error getting temperature and humidity data: %w", err)
			return response
		}
		response.TemperatureHumidityData = &TemperatureHumidityData{
			TemperatureCelsius: t,
			HumidityPercentage: h,
		}
	}

	return response
}

// Render is used to make this struct compatible with the go-chi webserver for writing
// the JSON response
func (g *GardenResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// AllGardensResponse is a simple struct being used to render and return a list of all Gardens
type AllGardensResponse struct {
	Gardens []*GardenResponse `json:"gardens"`
}

// NewAllGardensResponse will create an AllGardensResponse from a list of Gardens
func (gr GardensResource) NewAllGardensResponse(ctx context.Context, gardens []*pkg.Garden) *AllGardensResponse {
	gardenResponses := []*GardenResponse{}
	for _, g := range gardens {
		gardenResponses = append(gardenResponses, gr.NewGardenResponse(ctx, g))
	}
	return &AllGardensResponse{gardenResponses}
}

// Render is used to make this struct compatible with the go-chi webserver for writing
// the JSON response
func (pr *AllGardensResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}
