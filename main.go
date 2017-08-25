package main

import (
	"flag"
	"fmt"
	"github.com/noselasd/httpredirect/simplemux"
	"log"
	"net/http"
	"os"
	"runtime"
)

func accessLogger(l chan string) {
	for {
		log.Print(<-l)
	}
}

func httpLog(l chan string, w http.ResponseWriter, r *http.Request) {

	var remote string
	if len(r.Header["X-Forwarded-For"]) > 0 {
		remote = r.Header["X-Forwarded-For"][0]
	} else {
		remote = r.RemoteAddr
	}
	l <- fmt.Sprintf("%s %s %s", remote, r.Method, r.URL)
}

func httpWrapper(l chan string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "Go HTTP handler")
		handler.ServeHTTP(w, r)
		if *g_HttpLog {
			httpLog(l, w, r)
		}
	})
}

func redirect(w http.ResponseWriter, r *http.Request, matches []string) {
	http.Redirect(w, r, *g_RedirectTo, *g_RedirectStatus)

}

func serveHTTP(port string) {
	accessLogChan := make(chan string, 64)
	go accessLogger(accessLogChan)

	reHandler := simplemux.NewRegexpHandler()
	reHandler.AddRoute("^/", "", redirect)
	http.Handle("/", reHandler)

	var scheme string
	if *g_UseTLS {
		scheme = "https"
	} else {
		scheme = "http"
	}
	log.Printf("httpredirect(%s) listening at %s port %s\n", g_Version, scheme, port)

	var err error

	if *g_UseTLS {
		err = http.ListenAndServeTLS(":"+port, *g_TLSCert, *g_TLSKey, httpWrapper(accessLogChan, http.DefaultServeMux))
	} else {
		err = http.ListenAndServe(":"+port, httpWrapper(accessLogChan, http.DefaultServeMux))
	}
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}

var g_Port = flag.Int("port", 80, "port number to listen on")
var g_HttpLog = flag.Bool("httplog", true, "Whether to log HTTP requests to stdout")
var g_UseTLS = flag.Bool("usetls", false, "Use TLS(HTTPS) intead of plain HTTP")
var g_TLSCert = flag.String("tlscert", "tls.cert", "Path to TLS certificate file")
var g_TLSKey = flag.String("tlskey", "tls.key", "Path to TLS key file")
var g_Version = "DEVELOPMENT"
var g_RedirectTo = flag.String("target", "", "Target URL to redirect to")
var g_RedirectStatus = flag.Int("status", 307, "HTTP status code for the redirect")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n%s v.%s\n", os.Args[0], g_Version)
	}

	flag.Parse()
	if len(*g_RedirectTo) == 0 {
		fmt.Fprintln(os.Stderr, "target URL missing")
		flag.Usage()
		os.Exit(1)
	}

	runtime.GOMAXPROCS(3)

	serveHTTP(fmt.Sprintf("%d", *g_Port))
}
