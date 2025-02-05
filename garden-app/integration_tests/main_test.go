package integrationtests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/calvinmclean/automated-garden/garden-app/controller"
	"github.com/calvinmclean/automated-garden/garden-app/pkg"
	"github.com/calvinmclean/automated-garden/garden-app/pkg/action"
	"github.com/calvinmclean/automated-garden/garden-app/server"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	configFile = "testdata/config.yml"
)

var c *controller.Controller

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	serverConfig, controllerConfig := getConfigs(t)

	s, err := server.NewServer(serverConfig, true)
	require.NoError(t, err)

	c, err = controller.NewController(controllerConfig)
	require.NoError(t, err)

	go c.Start()
	go s.Start()

	defer c.Stop()
	defer s.Stop()

	time.Sleep(500 * time.Millisecond)

	t.Run("Garden", GardenTests)
	t.Run("Zone", ZoneTests)
}

func getConfigs(t *testing.T) (server.Config, controller.Config) {
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	require.NoError(t, err)

	var serverConfig server.Config
	err = viper.Unmarshal(&serverConfig)
	require.NoError(t, err)
	serverConfig.LogConfig.Level = logrus.DebugLevel.String()

	var controllerConfig controller.Config
	err = viper.Unmarshal(&controllerConfig)
	require.NoError(t, err)
	controllerConfig.LogConfig.Level = logrus.DebugLevel.String()

	return serverConfig, controllerConfig
}

func CreateGardenTest(t *testing.T) string {
	var g server.GardenResponse

	t.Run("CreateGarden", func(t *testing.T) {
		status, err := makeRequest(http.MethodPost, "/gardens", `{
			"name": "Test",
			"topic_prefix": "test",
			"max_zones": 3,
			"light_schedule": {
				"duration": "14h",
				"start_time": "22:00:00-07:00"
			},
			"temperature_humidity_sensor": true
		}`, &g)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusCreated, status)
	})

	return g.ID.String()
}

func GardenTests(t *testing.T) {
	gardenID := CreateGardenTest(t)

	t.Run("GetGarden", func(t *testing.T) {
		var g server.GardenResponse
		status, err := makeRequest(http.MethodGet, "/gardens/"+gardenID, http.NoBody, &g)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)

		assert.Equal(t, gardenID, g.ID.String())
		assert.Equal(t, uint(3), *g.MaxZones)
		assert.Equal(t, uint(0), g.NumZones)
		assert.Equal(t, uint(0), g.NumPlants)
	})
	t.Run("ExecuteStopAction", func(t *testing.T) {
		status, err := makeRequest(
			http.MethodPost,
			fmt.Sprintf("/gardens/%s/action", gardenID),
			action.GardenAction{Stop: &action.StopAction{}},
			&struct{}{},
		)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, status)

		time.Sleep(100 * time.Millisecond)

		c.AssertStopActions(t, 1)
	})
	t.Run("ExecuteStopAllAction", func(t *testing.T) {
		status, err := makeRequest(
			http.MethodPost,
			fmt.Sprintf("/gardens/%s/action", gardenID),
			action.GardenAction{Stop: &action.StopAction{All: true}},
			&struct{}{},
		)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, status)

		time.Sleep(100 * time.Millisecond)

		c.AssertStopAllActions(t, 1)
	})
	for _, state := range []pkg.LightState{pkg.LightStateOn, pkg.LightStateOff, pkg.LightStateToggle} {
		t.Run("ExecuteLightAction"+state.String(), func(t *testing.T) {
			status, err := makeRequest(
				http.MethodPost,
				fmt.Sprintf("/gardens/%s/action", gardenID),
				action.GardenAction{Light: &action.LightAction{State: state}},
				&struct{}{},
			)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusAccepted, status)

			time.Sleep(100 * time.Millisecond)

			c.AssertLightActions(t, action.LightAction{State: state})
		})
	}
	t.Run("ExecuteLightActionWithDelay", func(t *testing.T) {
		// Create new Garden with LightOnTime in the near future, so LightDelay will assume the light is currently off,
		// meaning adhoc action is going to be predictably delayed
		maxZones := uint(1)
		startTime := time.Now().In(time.Local).Add(1 * time.Second).Truncate(time.Second)
		newGarden := &server.GardenRequest{
			Garden: &pkg.Garden{
				Name:        "TestGarden",
				TopicPrefix: "test",
				MaxZones:    &maxZones,
				LightSchedule: &pkg.LightSchedule{
					Duration:  &pkg.Duration{Duration: 14 * time.Hour},
					StartTime: startTime.Format(pkg.LightTimeFormat),
				},
			},
		}

		var g server.GardenResponse
		status, err := makeRequest(http.MethodPost, "/gardens", newGarden, &g)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, status)

		// Execute light action with delay
		status, err = makeRequest(
			http.MethodPost,
			fmt.Sprintf("/gardens/%s/action", g.ID.String()),
			action.GardenAction{Light: &action.LightAction{
				State:       pkg.LightStateOff,
				ForDuration: &pkg.Duration{Duration: time.Second},
			}},
			&struct{}{},
		)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, status)

		time.Sleep(100 * time.Millisecond)

		// Make sure NextOnTime is correctly delayed
		var getG server.GardenResponse
		status, err = makeRequest(http.MethodGet, fmt.Sprintf("/gardens/%s", g.ID.String()), http.NoBody, &getG)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, startTime.Add(1*time.Second), getG.NextLightAction.Time.Local())

		time.Sleep(3 * time.Second)

		// Check for light action turning it off, plus adhoc schedule to turn it back on
		c.AssertLightActions(t,
			action.LightAction{State: pkg.LightStateOff, ForDuration: &pkg.Duration{Duration: time.Second}},
			action.LightAction{State: pkg.LightStateOn},
		)
	})
	t.Run("ChangeLightScheduleStartTimeResetsLightSchedule", func(t *testing.T) {
		// Reschedule Light to turn in in 1 second, for 1 second
		newStartTime := time.Now().Add(1 * time.Second).Truncate(time.Second)
		var g server.GardenResponse
		status, err := makeRequest(http.MethodPatch, "/gardens/"+gardenID, pkg.Garden{
			LightSchedule: &pkg.LightSchedule{
				StartTime: newStartTime.Format(pkg.LightTimeFormat),
				Duration:  &pkg.Duration{Duration: time.Second},
			},
		}, &g)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, newStartTime.Format(pkg.LightTimeFormat), g.LightSchedule.StartTime)

		time.Sleep(100 * time.Millisecond)

		// Make sure NextOnTime and state are changed
		var g2 server.GardenResponse
		status, err = makeRequest(http.MethodGet, "/gardens/"+gardenID, nil, &g2)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, newStartTime, g2.NextLightAction.Time.Truncate(time.Second).Local())
		assert.Equal(t, pkg.LightStateOn, g2.NextLightAction.State)

		time.Sleep(2 * time.Second)

		// Assert both LightActions
		c.AssertLightActions(t,
			action.LightAction{State: pkg.LightStateOn},
			action.LightAction{State: pkg.LightStateOff},
		)
	})
	t.Run("GetGardenToCheckInfluxDBData", func(t *testing.T) {
		var g server.GardenResponse
		status, err := makeRequest(http.MethodGet, "/gardens/"+gardenID, http.NoBody, &g)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)

		// The health status timing can be inconsistent, so it should be retried
		retries := 1
		for g.Health.Status != "UP" && retries <= 5 {
			time.Sleep(time.Duration(retries) * time.Second)

			status, err := makeRequest(http.MethodGet, "/gardens/"+gardenID, http.NoBody, &g)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, status)

			retries++
		}

		assert.Equal(t, "UP", g.Health.Status)
		assert.Equal(t, 50.0, g.TemperatureHumidityData.TemperatureCelsius)
		assert.Equal(t, 50.0, g.TemperatureHumidityData.HumidityPercentage)
	})
}

func CreateZoneTest(t *testing.T, gardenID, waterScheduleID string) string {
	var z server.ZoneResponse

	t.Run("CreateZone", func(t *testing.T) {
		status, err := makeRequest(http.MethodPost, fmt.Sprintf("/gardens/%s/zones", gardenID), fmt.Sprintf(`{
			"name": "Zone 1",
			"position": 0,
			"water_schedule_ids": ["%s"]
		}`, waterScheduleID), &z)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusCreated, status)
	})

	return z.ID.String()
}

func CreateWaterScheduleTest(t *testing.T) string {
	var ws server.WaterScheduleResponse

	t.Run("CreateWaterSchedule", func(t *testing.T) {
		status, err := makeRequest(http.MethodPost, "/water_schedules", `{
			"duration": "10s",
			"interval": "24h",
			"start_time": "2022-04-23T08:00:00-07:00"
		}`, &ws)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusCreated, status)
	})

	return ws.ID.String()
}

func ZoneTests(t *testing.T) {
	gardenID := CreateGardenTest(t)
	waterScheduleID := CreateWaterScheduleTest(t)
	zoneID := CreateZoneTest(t, gardenID, waterScheduleID)

	t.Run("ExecuteWaterAction", func(t *testing.T) {
		status, err := makeRequest(
			http.MethodPost,
			fmt.Sprintf("/gardens/%s/zones/%s/action", gardenID, zoneID),
			action.ZoneAction{Water: &action.WaterAction{
				Duration: &pkg.Duration{Duration: time.Second * 3},
			}},
			&struct{}{},
		)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, status)

		time.Sleep(100 * time.Millisecond)

		id, err := xid.FromString(zoneID)
		assert.NoError(t, err)
		c.AssertWaterActions(t, action.WaterMessage{
			Duration: 3000,
			ZoneID:   id,
			Position: 0,
		})
	})
	t.Run("CheckWateringHistory", func(t *testing.T) {
		// This test needs a few repeats to get a reliable pass, which is fine
		retries := 0

		var history server.ZoneWaterHistoryResponse
		for retries < 10 && history.Count < 1 {
			time.Sleep(300 * time.Millisecond)

			status, err := makeRequest(
				http.MethodGet,
				fmt.Sprintf("/gardens/%s/zones/%s/history", gardenID, zoneID),
				http.NoBody,
				&history,
			)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, status)
		}

		assert.Equal(t, 1, history.Count)
		assert.Equal(t, "3s", history.Average)
		assert.Equal(t, "3s", history.Total)
	})
	t.Run("ChangeWaterScheduleStartTimeResetsWaterSchedule", func(t *testing.T) {
		// Reschedule to Water in 2 second, for 1 second
		newStartTime := time.Now().Add(2 * time.Second).Truncate(time.Second)
		var ws server.WaterScheduleResponse
		status, err := makeRequest(http.MethodPatch, "/water_schedules/"+waterScheduleID, pkg.WaterSchedule{
			StartTime: &newStartTime,
			Duration:  &pkg.Duration{Duration: time.Second},
		}, &ws)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, newStartTime, ws.WaterSchedule.StartTime.Local())

		time.Sleep(100 * time.Millisecond)

		// Make sure NextWater is changed
		var z2 server.ZoneResponse
		status, err = makeRequest(http.MethodGet, fmt.Sprintf("/gardens/%s/zones/%s", gardenID, zoneID), http.NoBody, &z2)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, newStartTime, z2.NextWater.Time.Truncate(time.Second).Local())

		time.Sleep(3 * time.Second)

		// Assert WaterAction
		id, err := xid.FromString(zoneID)
		assert.NoError(t, err)
		c.AssertWaterActions(t,
			action.WaterMessage{
				Duration: 1000,
				ZoneID:   id,
				Position: 0,
			},
		)
	})
}

func makeRequest(method, path string, body, response interface{}) (int, error) {
	var reqBody io.Reader
	switch v := body.(type) {
	case nil:
	case string:
		reqBody = bytes.NewBuffer([]byte(v))
	case io.Reader:
		reqBody = v
	default:
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, "http://localhost:8080"+path, reqBody)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(data, response)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}
