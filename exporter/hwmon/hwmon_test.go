package hwmon

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestRapl(t *testing.T) {
	rapl, err := IntelRaplEnergy()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(rapl)
}

func TestCollector(t *testing.T) {
	col, _ := NewCollector(Conf{})
	prometheus.MustRegister(col)
	http.ListenAndServe(":8188", promhttp.Handler())
}
