package cmd

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/deorth-kku/go-common"
)

func curl(url string) error {
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	_, err = io.Copy(io.Discard, rsp.Body)
	if err != nil {
		return err
	}
	if rsp.StatusCode != http.StatusOK {
		return errors.New("unexpect return code " + rsp.Status)
	}
	return nil
}

func TestPprof(t *testing.T) {
	p := PprofPath{
		"goroutine": "/goroutine",
		"profile":   "/profile",
	}
	s := common.NewHttpServer()
	for path, h := range p.Handlers {
		s.HandleFunc(path, h)
	}
	go s.ListenAndServe(":12700")
	time.Sleep(100 * time.Microsecond)
	defer s.Shutdown(context.Background())
	err := curl("http://127.0.0.1:12700/goroutine?debug=1")
	if err != nil {
		t.Error(err)
		return
	}
	err = curl("http://127.0.0.1:12700/profile?seconds=1")
	if err != nil {
		t.Error(err)
		return
	}
}
