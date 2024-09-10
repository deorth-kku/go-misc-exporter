package common

import (
	"log/slog"
	"os"
	"os/signal"
)

func SignalsCallback(cb func(), once bool, sigs ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sigs...)
	go func() {
		for {
			sig := <-c
			slog.Debug("recived signal", "sig", sig)
			cb()
			if once {
				break
			}
		}
	}()
}
