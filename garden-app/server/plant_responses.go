package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/calvinmclean/automated-garden/garden-app/pkg"
)

// AllPlantsResponse is a simple struct being used to render and return a list of all Plants
type AllPlantsResponse struct {
	Plants []*PlantResponse `json:"plants"`
}

// NewAllPlantsResponse will create an AllPlantsResponse from a list of Plants
func (pr PlantsResource) NewAllPlantsResponse(ctx context.Context, plants []*pkg.Plant, garden *pkg.Garden) *AllPlantsResponse {
	plantResponses := []*PlantResponse{}
	for _, p := range plants {
		plantResponses = append(plantResponses, pr.NewPlantResponse(ctx, garden, p))
	}
	return &AllPlantsResponse{plantResponses}
}

// Render will take the map of Plants and convert it to a list for a more RESTy response
func (pr *AllPlantsResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// PlantResponse is used to represent a Plant in the response body with the additional Moisture data
// and hypermedia Links fields
type PlantResponse struct {
	*pkg.Plant
	Links []Link `json:"links,omitempty"`
}

// NewPlantResponse creates a self-referencing PlantResponse
func (pr PlantsResource) NewPlantResponse(_ context.Context, garden *pkg.Garden, plant *pkg.Plant, links ...Link) *PlantResponse {
	gardenPath := fmt.Sprintf("%s/%s", gardenBasePath, garden.ID)
	links = append(links,
		Link{
			"self",
			fmt.Sprintf("%s%s/%s", gardenPath, plantBasePath, plant.ID),
		},
		Link{
			"garden",
			gardenPath,
		},
		Link{
			"zone",
			fmt.Sprintf("%s%s/%s", gardenPath, zoneBasePath, plant.ZoneID),
		},
	)
	return &PlantResponse{
		plant,
		links,
	}
}

// Render is used to make this struct compatible with the go-chi webserver for writing
// the JSON response
func (p *PlantResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// PlantWaterHistoryResponse wraps a slice of WaterHistory structs plus some aggregate stats for an HTTP response
type PlantWaterHistoryResponse struct {
	History []pkg.WaterHistory `json:"history"`
	Count   int                `json:"count"`
	Average string             `json:"average"`
	Total   string             `json:"total"`
}
