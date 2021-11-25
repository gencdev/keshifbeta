package keshif

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	routes map[string]Route
}

type HostProxy struct{}

var proxyList = map[string]*httputil.ReverseProxy{}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host

	if proxyList[host] != nil {
		proxyList[host].ServeHTTP(w, r)
		return
	}

	for name, route := range p.routes {
		if route.Vhost == host {
			remoteUrl, err := url.Parse(fmt.Sprintf("http://%s:%s", route.Ip, route.Port))

			if err != nil {
				log.Println("target "+name+" parse failed.", err)
				break
			}

			proxy := httputil.NewSingleHostReverseProxy(remoteUrl)
			proxyList[host] = proxy
			proxy.ServeHTTP(w, r)
			return
		}
	}

	w.Write([]byte("403: Host forbidden " + host))
}

func StartProxy(routes map[string]Route, tlsCrt string, tlsKey string) {
	proxy := &Proxy{routes}
	http.Handle("/", proxy)

	isTls := tlsCrt != "" && tlsKey != ""
	port := ":80"

	if isTls {
		port = ":443"
	}

	server := &http.Server{Addr: port, Handler: proxy}

	go func() {
		if tlsCrt != "" && tlsKey != "" {
			if err := server.ListenAndServeTLS("./tls/localhost.crt", "./tls/localhost.key"); err != nil {
				fmt.Println("Keshif proxy terminated")
				log.Fatal(err)
			}
			return
		}

		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Keshif proxy terminated")
			log.Fatal(err)
		}
	}()
}
