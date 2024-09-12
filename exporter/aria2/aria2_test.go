package aria2

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/siku2/arigo"
)

const (
	rpcport = "6801"
	btport  = "6802"
	secret  = "test"
	wsrpc   = "ws://localhost:" + rpcport + "/jsonrpc"
)

func startAria2() {
	cmd := exec.Command("aria2c", "--rpc-secret", secret, "--rpc-listen-port", rpcport, "--listen-port", btport, "--dht-listen-port=", btport)
	cmd.Start()
	time.Sleep(time.Second / 2)
}

func TestAria2(t *testing.T) {
	startAria2()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cli, err := arigo.DialContext(ctx, wsrpc, secret)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	v, err := cli.GetVersion()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(v)
	err = cli.Shutdown()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestIterStruct(t *testing.T) {
	for k, v := range IterStructJson(arigo.Stats{}) {
		fmt.Println(k, v)
	}
}

func TestCollector(t *testing.T) {
	startAria2()
	col, err := NewCollector(Conf{
		Servers: []Server{{
			Rpc:     wsrpc,
			Secret:  secret,
			Timeout: 10,
		}},
	})
	if err != nil {
		t.Error(err)
		return
	}
	col.servers[0].AddURI([]string{"https://www.google.com"}, nil)
	err = cmd.TestCollectorThenClose(col)
	if err != nil {
		t.Error(err)
		return
	}
}
