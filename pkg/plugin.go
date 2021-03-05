package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dgraph-io/ristretto"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"gitlab.com/schoentoon/rs-tools/lib/ge"
	"gitlab.com/schoentoon/rs-tools/lib/ge/itemdb"
)

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ge := &ge.Ge{
		Client:    http.DefaultClient,
		UserAgent: "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:82.0) Gecko/20100101 Firefox/82.0",
	}
	f, err := os.OpenFile("/data/itemdb.ljson", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	idb, err := itemdb.NewFromReader(f)
	if err != nil {
		panic(err)
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1 * 1024 * 1024,  // 1MB
		MaxCost:     32 * 1024 * 1024, // 32MB
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	ds := &GeDataSource{
		im:     im,
		ge:     ge,
		search: idb,

		cache: cache,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/searchItems", ds.searchItems)

	httpResourceHandler := httpadapter.New(mux)

	return datasource.ServeOpts{
		CallResourceHandler: httpResourceHandler,
		QueryDataHandler:    ds,
	}
}

// GeDataSource provides a data source backed by the rs ge 'api'
type GeDataSource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager

	ge     ge.GeInterface
	search ge.SearchItemInterface

	cache *ristretto.Cache
}

// calculate the size in bytes so we know how much this would cost for the cache
func graphToSize(g *ge.Graph) int64 {
	size := 8 // the size of the ItemID

	// 24 is the size of time.Time and 4 is the size of int32
	// we just multiply this by the amount of entries in the graph
	size += ((24 + 4) * len(g.Graph))

	return int64(size)
}

func graphToTTL(now time.Time, g *ge.Graph) time.Duration {
	now = now.UTC()

	latest, _ := g.LatestPrice()

	// if the latest price update is from today we will cache till the end of today, which sadly seems
	// to happen rather randomly throughout the day, sometimes it happens 5-ish hours after midnight UTC
	// sometimes it happens 21 hours after midnight UTC. However it never updates twice per day, so once
	// it updated it should be safe to just cache it till the end of the day (UTC)
	if now.Year() == latest.Year() && now.Month() == latest.Month() && now.Day() == latest.Day() {
		// we could use time.Until here, but using Sub on latest + 1 day makes this function at least testable
		return latest.Add(time.Hour * 24).Sub(now)
	}

	// otherwise we're just going to cache it for 1 minute, would mostly prevent unnecessary calls to the
	// runescape ge api
	return time.Minute
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *GeDataSource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	type task struct {
		d backend.DataResponse
		q backend.DataQuery
	}

	var wg sync.WaitGroup
	wg.Add(len(req.Queries))
	ch := make(chan task, len(req.Queries))

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		go func(q backend.DataQuery) {
			ch <- task{d: td.query(&wg, ctx, q), q: q}
		}(q)
	}

	go func() {
		wg.Wait()
		//close(ch)
	}()

	for range req.Queries {
		//for task := range ch {
		task := <-ch
		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[task.q.RefID] = task.d
	}

	return response, nil
}

type queryModel struct {
	ItemID string `json:"itemID"`
}

func (td *GeDataSource) query(wg *sync.WaitGroup, ctx context.Context, query backend.DataQuery) backend.DataResponse {
	defer wg.Done()
	// Unmarshal the json into our queryModel
	var qm queryModel

	response := backend.DataResponse{}

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	id, err := strconv.ParseInt(qm.ItemID, 10, 64)
	if err != nil {
		response.Error = err
		return response
	}

	item, err := td.search.GetItem(id)
	if err != nil {
		response.Error = err
		return response
	}

	var graph *ge.Graph

	val, found := td.cache.Get(id)
	if !found {
		err := backoff.Retry(func() error {
			graph, err = td.ge.PriceGraph(id)
			return err
		}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 3))
		if err != nil {
			response.Error = err
			return response
		}

		td.cache.SetWithTTL(id, graph, graphToSize(graph), graphToTTL(time.Now(), graph))
	} else {
		graph = val.(*ge.Graph)
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
	frame.Fields = append(frame.Fields, data.NewField(item.Name, nil, prices))

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
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

func (td *GeDataSource) searchItems(w http.ResponseWriter, req *http.Request) {
	log.DefaultLogger.Debug("Received resource call", "url", req.URL.String(), "method", req.Method)
	if req.Method != http.MethodPost {
		return
	}

	r := struct {
		Query string `json:"query"`
	}{}

	err := json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.DefaultLogger.Debug(fmt.Sprintf("Searching for '%s'", r.Query))
	results, err := td.search.SearchItems(r.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.DefaultLogger.Debug(fmt.Sprintf("results: %#v", results))

	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		log.DefaultLogger.Error("Json encoding error.. ", err)
	}
}
