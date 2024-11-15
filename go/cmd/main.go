package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"polaris/truffle/pkg/client"
	"polaris/truffle/pkg/common"
	"polaris/truffle/pkg/metrics"
	"polaris/truffle/pkg/server"
	"polaris/truffle/pkg/utils"
	"time"
)

const IncomingPath = "/hello"
const OutgoingPath = "/outgoing"

var (
	ns, label, field             string
	incomingProxy, outgoingProxy bool
)

func init() {

	flag.StringVar(&ns, "namespace", "", "namespace")
	flag.StringVar(&label, "l", "", "Label selector")
	flag.StringVar(&field, "f", "", "Field selector")
	flag.BoolVar(&common.Debug, "debug", false, "Debug log level")
	flag.BoolVar(&common.Trace, "trace", false, "Trace log level")
	flag.StringVar(&common.IncomingPodPort, "incoming-port", "8080", "pod incoming port")
	flag.StringVar(&common.AwsAccessKey, "access-key", "", "AWS access key")
	flag.StringVar(&common.AwsSecretKey, "secret-key", "", "AWS secret key")
	flag.StringVar(&common.ComMode, "comm-mode", "TRUFFLE", "Communication mode: TRUFFLE S3 KVS")
	flag.StringVar(&common.RedisIP, "redis-ip", "localhost", "Redis IP")
	flag.StringVar(&common.RedisPwd, "redis-pwd", "", "Redis Password")
	flag.BoolVar(&incomingProxy, "incoming-proxy", false, "Use reverse proxy for incoming requests")
	flag.BoolVar(&outgoingProxy, "outgoing-proxy", false, "Use reverse proxy for outgoing requests")
	flag.Parse()
	client.Pool = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:6379", common.RedisIP),
		Password:     common.RedisPwd, // no password set
		DB:           0,               // use default DB
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	})
	if common.Trace {
		common.Debug = true
	}

	logWriter := new(utils.LogWriter)
	common.DebugLog = log.New(logWriter, "[DEBUG] ", 0)
	common.TraceLog = log.New(logWriter, "[TRACE] ", 0)

	log.SetFlags(0)
	log.SetOutput(logWriter)

}

func main() {
	if common.Debug {
		common.DebugLog.Printf("This is something the label set %s", label)
	}
	go metrics.StartPodMetrics()
	r := mux.NewRouter()
	// Returns a proxy for the target url.
	proxy, err := server.NewProxy("http://localhost:8080")
	if err != nil {
		panic(err)
	}
	if r.HandleFunc(IncomingPath, server.IncomingHandler()); incomingProxy {
		r.HandleFunc(IncomingPath, server.ProxyIncomingHandler(proxy))
	}
	if r.HandleFunc(OutgoingPath, server.OutgoingHandler()); outgoingProxy {
		r.HandleFunc(OutgoingPath, server.ProxyOutgoingHandler(proxy))
	}
	//r.HandleFunc("/upload", api.Handle())
	err = http.ListenAndServe(":8888", r)
	if err != nil {
		return
	}

}
