package netatmo

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

const minTemperatureInterval = 72 * time.Hour

// GetAverageHighTemperature returns the average daily high temperature between the given time and the end of
// yesterday (since daily high can be misleading if queried mid-day)
func (c *Client) GetAverageHighTemperature(since time.Duration) (float32, error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("avg_temp_%d", now.Unix())
	cachedData, found := c.responseCache.Get(cacheKey)
	if found {
		return cachedData.(float32), nil
	}

	// Time to check since must always be at least 3 days
	if since < minTemperatureInterval {
		since = minTemperatureInterval
	}

	beginDate := now.Add(-since).Truncate(time.Hour)
	beginDate = time.Date(beginDate.Year(), beginDate.Month(), beginDate.Day()-1, 23, 59, 59, 0, time.Local)
	// Since we are looking at daily max temp, get time all the way to very end of yesterday
	endDate := time.Date(now.Year(), now.Month(), now.Day()-1, 23, 59, 59, 0, time.Local)

	temperatureData, err := c.getMeasure("max_temp", "1day", beginDate, &endDate)
	if err != nil {
		return 0, err
	}

	c.responseCache.Set(cacheKey, temperatureData.Average(), cache.DefaultExpiration)

	return temperatureData.Average(), nil
}
