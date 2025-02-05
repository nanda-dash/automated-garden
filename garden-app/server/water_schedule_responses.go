package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/calvinmclean/automated-garden/garden-app/pkg"
	"github.com/calvinmclean/automated-garden/garden-app/worker"
	"github.com/rs/xid"
)

// AllWaterSchedulesResponse is a simple struct being used to render and return a list of all WaterSchedules
type AllWaterSchedulesResponse struct {
	WaterSchedules []*WaterScheduleResponse `json:"water_schedules"`
}

// NewAllWaterSchedulesResponse will create an AllWaterSchedulesResponse from a list of WaterSchedules
func (wsr WaterSchedulesResource) NewAllWaterSchedulesResponse(ctx context.Context, waterschedules []*pkg.WaterSchedule, excludeWeatherData bool) *AllWaterSchedulesResponse {
	waterscheduleResponses := []*WaterScheduleResponse{}
	for _, ws := range waterschedules {
		waterscheduleResponses = append(waterscheduleResponses, wsr.NewWaterScheduleResponse(ctx, ws, excludeWeatherData))
	}
	return &AllWaterSchedulesResponse{waterscheduleResponses}
}

// Render ...
func (zr *AllWaterSchedulesResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// WaterScheduleResponse is used to represent a WaterSchedule in the response body with the additional Moisture data
// and hypermedia Links fields
type WaterScheduleResponse struct {
	*pkg.WaterSchedule
	WeatherData *WeatherData     `json:"weather_data,omitempty"`
	NextWater   NextWaterDetails `json:"next_water,omitempty"`
	Links       []Link           `json:"links,omitempty"`
}

// NextWaterDetails has information about the next time this WaterSchedule will be used
type NextWaterDetails struct {
	Time            *time.Time `json:"time,omitempty"`
	Duration        string     `json:"duration,omitempty"`
	WaterScheduleID *xid.ID    `json:"water_schedule_id,omitempty"`
	Message         string     `json:"message,omitempty"`
}

// GetNextWaterDetails returns the NextWaterDetails for the WaterSchedule
func GetNextWaterDetails(ws *pkg.WaterSchedule, worker *worker.Worker, excludeWeatherData bool) NextWaterDetails {
	result := NextWaterDetails{
		Time:     worker.GetNextWaterTime(ws),
		Duration: ws.Duration.Duration.String(),
	}

	if ws.HasWeatherControl() && !excludeWeatherData {
		wd, hadErr := worker.ScaleWateringDuration(ws)
		if hadErr {
			result.Message = "error impacted duration scaling"
		}

		result.Duration = time.Duration(wd).String()
	}

	return result
}

// NewWaterScheduleResponse creates a self-referencing WaterScheduleResponse
func (wsr WaterSchedulesResource) NewWaterScheduleResponse(ctx context.Context, ws *pkg.WaterSchedule, excludeWeatherData bool, links ...Link) *WaterScheduleResponse {
	response := &WaterScheduleResponse{
		WaterSchedule: ws,
		Links:         links,
	}

	response.Links = append(response.Links,
		Link{
			"self",
			fmt.Sprintf("%s/%s", waterScheduleBasePath, ws.ID),
		},
	)

	if ws.HasWeatherControl() && !ws.EndDated() && !excludeWeatherData {
		response.WeatherData = getWeatherData(ctx, ws, wsr.storageClient)
	}

	if !ws.EndDated() {
		response.NextWater = GetNextWaterDetails(ws, wsr.worker, excludeWeatherData)
	}

	return response
}

// Render is used to make this struct compatible with the go-chi webserver for writing
// the JSON response
func (z *WaterScheduleResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}
