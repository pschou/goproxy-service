// Original source derived from:
// https://gist.github.com/Ompluscator/572e474ee054d72259ecffd010fec630/

package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var (
	about = `GoLang Proxy Redirect Service

This utility is intended to listen for GoLang proxy request and redirect them
to the proper git project for handling such requests.  This is used for
building against private projects and not having to expose the GitProject to
the public domain to use the git mod commands.

The Syntax of the config.yml:
| # exact matches for replacing a request to a target (ie: locally hosted)
| modules:
|   company.com/package-a: gitlab.com/pkg-a
|   company.com/package-b: gitlab.com/pkg-b
| 
| # default git credentials to use
| git-token: AAAAAAAAAABBBBBBBBBBBBBCCCCCCCCCCDDDDDDD
| git-url: https://gitlab.com
| 
| regexp:
| - match: "mytest.domain.A/([^/*])"
|   replace: "another.domain/a/$1"
|   git-token: AAAAAAAAAABBBBBBBBBBBBBCCCCCCCCCCDDDDDDD
|   git-url: https://another.domain
|   # alternate domain can be substituted with a regexp match and replace
| - match: "github.com.*"
|   git-token: AAAAAAAAAABBBBBBBBBBBBBCCCCCCCCCCDDDDDDD
|   git-url: https://github.com
|   # without a replace, the original url is used with the provided token
`

	listen         = flag.String("listen", ":8080", "Where to listen to incoming connections (example 1.2.3.4:8080)")
	enableTLS      = flag.Bool("tls", false, "Enforce TLS secure transport on incoming connections")
	verbose        = flag.Bool("verbose", false, "Turn on verbose")
	compileVersion = "SELF BUILT"
	usage          = "[options]"
)

func main() {
	flag.Parse()
	loadTLS()
	loadConfig()

	// setup server for proxying packages
	router := mux.NewRouter()
	router.HandleFunc("/{module:.+}/@v/list", list).Methods(http.MethodGet)
	router.HandleFunc("/{module:.+}/@v/{version}.info", version).Methods(http.MethodGet)
	router.HandleFunc("/{module:.+}/@v/{version}.mod", mod).Methods(http.MethodGet)
	router.HandleFunc("/{module:.+}/@v/{version}.zip", archive).Methods(http.MethodGet)
	router.HandleFunc("/{module:.+}/@v/{version}.sum", sum).Methods(http.MethodGet)
	router.HandleFunc("/{module:.+}/@latest", latest).Methods(http.MethodGet)

	// setup server for summing packages
	router.HandleFunc("/lookup/{module:.+}@{version}", sum).Methods(http.MethodGet)

	http.Handle("/", router)
	// Configure the go HTTP server
	server := &http.Server{
		Addr:           *listen,
		TLSConfig:      tlsConfig,
		ReadTimeout:    10 * time.Hour,
		WriteTimeout:   10 * time.Hour,
		MaxHeaderBytes: 1 << 20,
	}

	if *enableTLS {
		log.Println("Listening with HTTPS on", *listen)
		log.Fatal(server.ListenAndServeTLS(*certFile, *keyFile))
	} else {
		log.Println("Listening with HTTP on", *listen)
		log.Fatal(server.ListenAndServe())
	}
}
