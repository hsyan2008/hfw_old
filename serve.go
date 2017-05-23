package hfw

import (
	"net/http"
	"time"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/hsyan2008/go-logger/logger"
)

func startServe() {

	s := &http.Server{
		Addr: ":" + Config.Server.Port,
		// Handler:      controllers,
		ReadTimeout:  time.Duration(Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(Config.Server.WriteTimeout) * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}
	//kill -USR2 pid 来重启
	err := gracehttp.Serve(s)
	// err:= s.ListenAndServe()

	if err != nil {
		logger.Fatal("ListenAndServe: ", err)
	}
}
