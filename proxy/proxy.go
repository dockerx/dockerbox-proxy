package proxy

import (
	"io"
	"log"
	"net"
	"strings"
	"net/http"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/testutils"
	"github.com/dockerx/dockerbox-proxy/backend"
)


func isWebsocket(req *http.Request) bool {
	conn_hdr := ""
	conn_hdrs := req.Header["Connection"]
	if len(conn_hdrs) > 0 {
		conn_hdr = conn_hdrs[0]
	}
	upgrade_websocket := false
	if strings.ToLower(conn_hdr) == "upgrade" {
		upgrade_hdrs := req.Header["Upgrade"]
		if len(upgrade_hdrs) > 0 {
			upgrade_websocket = (strings.ToLower(upgrade_hdrs[0]) == "websocket")
		}
	}
	return upgrade_websocket
}

func websocketProxy(target string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws_socket, err := net.Dial("tcp", target)
		if err != nil {
			http.Error(w, "Backend is not available", 502)
			log.Printf("Error dialing websocket backend %s: %v", target, err)
			return
		}
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Error: creating proxy", 500)
			return
		}
		nc, _, err := hj.Hijack()
		if err != nil {
			log.Printf("Hijack error: %v", err)
			return
		}
		defer nc.Close()
		defer ws_socket.Close()
		err = r.Write(ws_socket)
		if err != nil {
			log.Printf("Error copying request to target: %v", err)
			return
		}
		errc := make(chan error, 2)
		cp := func(dst io.Writer, src io.Reader) {
			_, err := io.Copy(dst, src)
			errc <- err
		}
		go cp(ws_socket, nc)
		go cp(nc, ws_socket)
		<-errc
	})
}

func proxyHandler() http.Handler {
	proxyHandleFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := backend.GetTarget(r)
		if target == "" {
			log.Println()
			http.Error(w, "Bad Gateway", 502)
			return
		}
		//WebSocket Connection
		if isWebsocket(r) {
			log.Println("Initializing WS proxy : ", target)
			ws_proxy := websocketProxy(target)
			ws_proxy.ServeHTTP(w, r)
			return
		}
		//All other HTTP requests
		httpProxy, _ := forward.New(forward.PassHostHeader(true))
		r.URL = testutils.ParseURI("http://"+target)
		httpProxy.ServeHTTP(w, r)
	})
	return proxyHandleFunc
}

func StartProxy() {
	go func() {
		proxyHandleFunc := proxyHandler()
		s := &http.Server{
			Addr:           ":80",
			Handler:        proxyHandleFunc,
		}
		err := s.ListenAndServe()
		if err != nil {
			panic(err)
		}
		log.Println("Shutting down proxy")
	}()
}
