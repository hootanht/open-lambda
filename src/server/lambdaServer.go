package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/open-lambda/open-lambda/ol/common"
	"github.com/open-lambda/open-lambda/ol/lambda"
)

// LambdaServer is a worker server that listens to run lambda requests and forward
// these requests to its sandboxes.
type LambdaServer struct {
	lambdaMgr *lambda.LambdaMgr
}

// getUrlComponents parses request URL into its "/" delimated components
func getUrlComponents(r *http.Request) []string {
	path := r.URL.Path

	// trim prefix
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	// trim trailing "/"
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	components := strings.Split(path, "/")
	return components
}

// RunLambda expects POST requests like this:
//
// curl localhost:PORT/run/<lambda-name>
// curl -X POST localhost:PORT/run/<lambda-name> -d '{}'
// ...
func (s *LambdaServer) Invoke(w http.ResponseWriter, r *http.Request) {
	t := common.T0("web-request")
	defer t.T1()

	log.Printf("Received request to %s\n", r.URL.Path)

	// components represent run[0]/<name_of_sandbox>[1]/<extra_things>...
	// ergo we want [1] for name of sandbox
	urlParts := getUrlComponents(r)
	if len(urlParts) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expected invocation format: /run/<lambda-name>"))
		return
	}

	img := urlParts[1]
	s.lambdaMgr.Get(img).Invoke(w, r)
}

func (s *LambdaServer) Debug(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(s.lambdaMgr.Debug()))
}

func (s *LambdaServer) cleanup() {
	s.lambdaMgr.Cleanup()
}

// NewLambdaServer creates a server based on the passed config."
func NewLambdaServer() (*LambdaServer, error) {
	log.Printf("Starting new lambda server")

	lambdaMgr, err := lambda.NewLambdaMgr()
	if err != nil {
		return nil, err
	}

	server := &LambdaServer{
		lambdaMgr: lambdaMgr,
	}

	log.Printf("Setups Handlers")
	port := fmt.Sprintf(":%s", common.Conf.Worker_port)
	http.HandleFunc(RUN_PATH, server.Invoke)
	http.HandleFunc(DEBUG_PATH, server.Debug)

	log.Printf("Execute handler by POSTing to localhost%s%s%s\n", port, RUN_PATH, "<lambda>")
	log.Printf("Get status by sending request to localhost%s%s\n", port, STATUS_PATH)

	return server, nil
}
