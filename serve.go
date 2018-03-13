package hfw

import (
	"crypto/tls"
	"net/http"
	"time"

	"golang.org/x/net/http2"

	"github.com/hsyan2008/go-logger/logger"
	"github.com/hsyan2008/grace/gracehttp"
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

//支持https、grace
func startHttpsServe(certFile, keyFile string) {

	var err error
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logger.Fatal("load cert file error:", err)
	}

	s := &http.Server{
		Addr: ":" + Config.Server.Port,
		// Handler:      controllers,
		ReadTimeout:  time.Duration(Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(Config.Server.WriteTimeout) * time.Second,
		// MaxHeaderBytes: 1 << 20,
		TLSConfig: &tls.Config{
			NextProtos: []string{"http/1.1", "h2"},
			Certificates: []tls.Certificate{
				cert,
			},
		},
	}

	//kill -USR2 pid 来重启
	err = gracehttp.Serve(s)

	if err != nil {
		logger.Fatal("ListenAndServeTls: ", err)
	}
}

//支持http2，但不支持grace
func startHttpsServe2(certFile, keyFile string) {

	var err error

	s := &http.Server{
		Addr: ":" + Config.Server.Port,
		// Handler:      controllers,
		ReadTimeout:  time.Duration(Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(Config.Server.WriteTimeout) * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}

	_ = http2.ConfigureServer(s, nil)

	err = s.ListenAndServeTLS(certFile, keyFile)

	if err != nil {
		logger.Fatal("ListenAndServeTls: ", err)
	}
}
