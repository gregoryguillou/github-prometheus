package main

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestParseMetrics(t *testing.T) {
	v, err := ParseMetrics("config-test2.yml")
	if err != nil {
		t.Error(err)
	}
	if len(v.Metrics) < 1 {
		t.Error("should return several metrics")
	}
}

//go:embed config-test1.json
var configTest1 []byte

func TestParseOutput(t *testing.T) {

	listData, err := mapToList(configTest1, "organization.repositories.edges")
	if err != nil {
		t.Error(err)
	}
	labels := []string{"name"}
	gaugevec := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "github_issues_total",
		Help: "number of github PR based on ",
	}, labels)

	err = setGauges(gaugevec, listData, []NamedValue{{Name: "name", Value: "node.name"}}, "node.issues.totalCount")
	if err != nil {
		t.Error(err)
	}
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := promhttp.Handler()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), `github_issues_total{name="demo1"} 321`) {
		t.Errorf("metric should contain: github_issues_total{name=\"demo1\"} 321\n, yet:\n%s", rr.Body.String())
	}
}
