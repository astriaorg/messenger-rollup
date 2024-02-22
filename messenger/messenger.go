package messenger

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	astriaGrpc "buf.build/gen/go/astria/execution-apis/grpc/go/astria/execution/v1alpha2/executionv1alpha2grpc"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

type Config struct {
	SequencerRPC string `env:"SEQUENCER_RPC, default=http://localhost:26657"`
	ConductorRPC string `env:"CONDUCTOR_RPC, default=:50051"`
	RESTApiPort  string `env:"RESTAPI_PORT, default=:8080"`
	RollupId     string `env:"ROLLUP_ID, default=messenger-rollup"`
}

// App is the main application struct, containing all the necessary components.
type App struct {
	executionRPC    string
	sequencerRPC    string
	sequencerClient SequencerClient
	restRouter      *mux.Router
	restAddr        string
	messenger       *Messenger
	rollupId        []byte
}

func NewApp(cfg Config) *App {
	log.Debugf("creating new messenger with config: %v", cfg)

	m := NewMessenger()
	router := mux.NewRouter()

	rollupId := sha256.Sum256([]byte(cfg.RollupId))

	return &App{
		executionRPC:    cfg.ConductorRPC,
		sequencerRPC:    cfg.SequencerRPC,
		sequencerClient: *NewSequencerClient(cfg.SequencerRPC, rollupId[:]),
		restRouter:      router,
		restAddr:        cfg.RESTApiPort,
		messenger:       m,
		rollupId:        rollupId[:],
	}
}

// makeExecutionServer creates a new ExecutionServiceServer.
func (a *App) makeExecutionServer() *ExecutionServiceServerV1Alpha2 {
	return NewExecutionServiceServerV1Alpha2(a.messenger, a.rollupId)
}

// setupRestRoutes sets up the routes for the REST API.
func (a *App) setupRestRoutes() {
	a.restRouter.HandleFunc("/block/{height}", a.getBlock).Methods("GET")
	a.restRouter.HandleFunc("/message", a.postMessage).Methods("POST")
}

// makeRestServer creates a new HTTP server for the REST API.
func (a *App) makeRestServer() *http.Server {
	return &http.Server{
		Addr:    a.restAddr,
		Handler: a.restRouter,
	}
}

func (a *App) getBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	heightStr, ok := vars["height"]
	if !ok {
		log.Errorf("error getting height from request\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		log.Errorf("error converting height to int: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debugf("getting block %d\n", height)
	block, err := a.messenger.GetSingleBlock(uint32(height))
	if err != nil {
		log.Errorf("error getting block: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	blockJson, err := json.Marshal(block)
	if err != nil {
		log.Errorf("error marshalling block: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(blockJson)
}

func (a *App) postMessage(w http.ResponseWriter, r *http.Request) {
	var tx Transaction
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		log.Errorf("error decoding transaction: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resp, err := a.sequencerClient.SendMessage(tx)
	if err != nil {
		log.Errorf("error sending message: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Debugf("postMessage response: %v\n", resp.Log)
}

func (a *App) Run() {
	// run execution api
	go func() {
		server := a.makeExecutionServer()
		lis, err := net.Listen("tcp", a.executionRPC)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		astriaGrpc.RegisterExecutionServiceServer(grpcServer, server)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// run rest api server
	a.setupRestRoutes()
	server := a.makeRestServer()
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Errorf("rest api server closed\n")
	} else if err != nil {
		log.Errorf("error listening for rest api server: %s\n", err)
	}
}
