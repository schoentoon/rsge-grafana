package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"gitlab.com/schoentoon/rs-tools/lib/ge"
)

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ds := &GeDataSource{
		im: im,
		ge: &ge.Ge{Client: http.DefaultClient},
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

// GeDataSource provides a data source backed by the rs ge 'api'
type GeDataSource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager

	ge ge.GeInterface
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *GeDataSource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := td.query(ctx, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	ItemID int64 `json:"itemID"`
}

func (td *GeDataSource) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	// Unmarshal the json into our queryModel
	var qm queryModel

	response := backend.DataResponse{}

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	graph, err := td.ge.PriceGraph(qm.ItemID)
	if err != nil {
		response.Error = err
		return response
	}

	// create data frame response
	frame := data.NewFrame("response")

	times := make([]time.Time, 0, len(graph.Graph))
	prices := make([]int32, 0, len(graph.Graph))

	// we first filter out all the times and sort them
	for when := range graph.Graph {
		if when.After(query.TimeRange.From) && when.Before(query.TimeRange.To) {
			times = append(times, when)
		}
	}

	sort.SliceStable(times, func(i, j int) bool { return times[i].Unix() < times[j].Unix() })

	// after that we iterate over the now sorted times and build the second array with prices associated with it
	for _, when := range times {
		prices = append(prices, int32(graph.Graph[when]))
	}

	// add the time dimension
	frame.Fields = append(frame.Fields, data.NewField("time", nil, times))

	// add values
	frame.Fields = append(frame.Fields, data.NewField("values", nil, prices))

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *GeDataSource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// TODO
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

type instanceSettings struct {
	httpClient *http.Client
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &instanceSettings{
		httpClient: &http.Client{},
	}, nil
}

func (s *instanceSettings) Dispose() {
	// Called before creating a new instance to allow plugin authors
	// to cleanup.
}
