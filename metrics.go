package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/errgroup"
)

type Query struct {
	Query string `json:"query"`
}

// query run a graphql query on Github
func query(ctx context.Context, endpoint, bearer, query string) ([]byte, error) {
	q := Query{
		Query: "query " + query,
	}
	payload, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "bearer "+bearer)
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// mapToList returns the slice of data for a listkey
func mapToList(content []byte, listKey string) ([]map[string]interface{}, error) {
	result := map[string]interface{}{}
	err := json.Unmarshal(content, &result)
	if err != nil {
		return nil, err
	}
	keys := strings.Split(listKey, ".")
	values := []string{"data"}
	if len(keys) > 1 {
		values = append(values, keys[:len(keys)-1]...)
	}
	dataContent := result
	for _, key := range values {
		data, ok := dataContent[key]
		if !ok {
			return nil, fmt.Errorf("no data key")
		}
		dataContent, ok = data.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("no data content")
		}
	}
	l, ok := dataContent[keys[len(keys)-1]]
	if !ok {
		return nil, fmt.Errorf("no list")
	}
	listContent, ok := l.([]interface{})
	if !ok {
		return nil, fmt.Errorf("no list content")
	}
	output := []map[string]interface{}{}
	for _, iter := range listContent {
		iterContent, ok := iter.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("empty iteration")
		}
		output = append(output, iterContent)
	}
	return output, nil
}

// setGauges fill gaugevec with query based on labels
func setGauges(gaugevec *prometheus.GaugeVec, data []map[string]interface{}, labels []NamedValue, value interface{}) error {
	capturedValue := 1.0
	switch v := value.(type) {
	case int:
		capturedValue = float64(v)
	case int64:
		capturedValue = float64(v)
	case float64:
		capturedValue = v
	case string:
	default:
		return fmt.Errorf("unexpected type %T", value)
	}
	for _, iter := range data {
		newLabels := prometheus.Labels{}
		for _, w := range labels {
			internalKeys := strings.Split(w.Value, ".")
			internalDataContent := iter
			for _, key := range internalKeys[:len(internalKeys)-1] {
				internalData, ok := internalDataContent[key]
				if !ok {
					return fmt.Errorf("no data key")
				}
				internalDataContent, ok = internalData.(map[string]interface{})
				if !ok {
					return fmt.Errorf("no data content")
				}
			}
			newLabels[w.Name] = fmt.Sprintf("%v", internalDataContent[internalKeys[len(internalKeys)-1]])
		}
		if v, ok := value.(string); ok {
			internalKeys := strings.Split(v, ".")
			internalDataContent := iter
			for _, key := range internalKeys[:len(internalKeys)-1] {
				internalData, ok := internalDataContent[key]
				if !ok {
					return fmt.Errorf("no data key")
				}
				internalDataContent, ok = internalData.(map[string]interface{})
				if !ok {
					return fmt.Errorf("no data content")
				}
			}
			s := fmt.Sprintf("%v", internalDataContent[internalKeys[len(internalKeys)-1]])
			var err error
			capturedValue, err = strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
		}
		gauge, _ := gaugevec.GetMetricWith(newLabels)
		gauge.Set(capturedValue)
	}
	return nil
}

// runMetric run one metric and set the associated gauges
func runMetric(ctx context.Context, metric Metric, gaugevec *prometheus.GaugeVec) error {
	currentToken := metric.Bearer.PersonalAccessToken
	if currentToken == nil && metric.Bearer.Endpoint != nil {
		t, err := token(ctx, *metric.Bearer.Endpoint)
		if err != nil {
			return err
		}
		currentToken = &t.Token
	}
	content, err := query(ctx, metric.Endpoint, *currentToken, metric.Query)
	if err != nil {
		return err
	}
	listData, err := mapToList(content, metric.List)
	if err != nil {
		return err
	}
	return setGauges(gaugevec, listData, metric.Labels, metric.Value)
}

// runMetrics schedule the running of all the metrics and stop if one fails
func runMetrics(ctx context.Context, cancel context.CancelFunc, metrics Metrics) func() error {
	return func() error {
		defer cancel()
		g, ctx := errgroup.WithContext(ctx)
		ctx, c := context.WithCancel(ctx)
		defer c()

		for _, metric := range metrics.Metrics {
			labels := []string{}
			for _, w := range metric.Labels {
				labels = append(labels, w.Name)
			}
			gaugevec := promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: metric.Name,
				Help: metric.Help,
			}, labels)

			g.Go(func(m Metric, c context.CancelFunc) func() error {
				return func() error {
					defer c()
					err := runMetric(ctx, m, gaugevec)
					if err != nil {
						fmt.Println("error runQuery:", err)
						return err
					}
					tick := time.NewTicker(time.Second * 30)
					for {
						select {
						case <-tick.C:
							err := runMetric(ctx, m, gaugevec)
							if err != nil {
								fmt.Println("error runQuery:", err)
								return err
							}
						case <-ctx.Done():
							return nil
						}
					}
				}
			}(metric, c))
		}
		return g.Wait()
	}
}
