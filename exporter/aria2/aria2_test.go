package aria2

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/deorth-kku/aria2rpc-go"
	"github.com/deorth-kku/go-misc-exporter/cmd"
)

var _ cmd.Collector = new(collector)

const (
	rpcport = "6801"
	btport  = "6802"
	secret  = "test"
	wsrpc   = "ws://localhost:" + rpcport + "/jsonrpc"
)

func startAria2() {
	cmd := exec.Command("aria2c",
		"--enable-rpc",
		"--rpc-secret="+secret,
		"--rpc-listen-port="+rpcport,
		"--listen-port="+btport,
		"--dht-listen-port="+btport,
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Start()
	time.Sleep(time.Second / 2)
}

func TestAria2(t *testing.T) {
	startAria2()
	ctx := t.Context()
	cli, err := aria2rpc.New(ctx, wsrpc, aria2rpc.WithSecret(secret))
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	v, err := cli.GetVersion(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(v)
	_, err = cli.Shutdown(ctx)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestIterStruct(t *testing.T) {
	for k, v := range IterStructJson(aria2rpc.GlobalStat{}) {
		fmt.Println(k, v)
	}
}

func TestCollector(t *testing.T) {
	startAria2()
	ctx := t.Context()
	col, err := NewCollector(Conf{
		Servers: []ServerConf{{
			Rpc:     wsrpc,
			Secret:  secret,
			Timeout: 10,
		}},
	})
	if err != nil {
		t.Error(err)
		return
	}
	col.servers[0].AddURI(ctx, []string{"https://www.google.com"}, nil, nil)
	err = cmd.TestCollector(col)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = col.servers[0].Shutdown(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(10 * time.Second)
	startAria2()
	time.Sleep(1 * time.Second)
	err = cmd.TestCollectorThenClose(col)
	if err != nil {
		t.Error(err)
		return
	}
}
