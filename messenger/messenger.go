package messenger

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	astriaGrpc "buf.build/gen/go/astria/astria/grpc/go/astria/execution/v1alpha2/executionv1alpha2grpc"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func (a *App) getBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	heightStr, ok := vars["height"]
	if !ok {
		fmt.Printf("error getting height from request\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		fmt.Printf("error converting height to int: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("getting block %d\n", height)
	block, err := a.messenger.GetSingleBlock(uint32(height))
	if err != nil {
		fmt.Printf("error getting block: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	blockJson, err := json.Marshal(block)
	if err != nil {
		fmt.Printf("error marshalling block: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(blockJson)
}

func (a *App) postMessage(w http.ResponseWriter, r *http.Request) {
	var tx Transaction
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		fmt.Printf("error decoding transaction: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resp, err := a.sequencerClient.SendMessage(tx)
	if err != nil {
		fmt.Printf("error sending message: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	println(resp.Log)
}

type App struct {
	executionAddr   string
	sequencerAddr   string
	sequencerClient SequencerClient
	restRouter      *mux.Router
	restAddr        string
	messenger       *Messenger
}

func NewApp(executionAddr string, sequencerAddr string, restAddr string) *App {
	m := NewMessenger()

	router := mux.NewRouter()
	return &App{
		executionAddr:   executionAddr,
		sequencerAddr:   sequencerAddr,
		sequencerClient: *NewSequencerClient(sequencerAddr),
		restRouter:      router,
		restAddr:        restAddr,
		messenger:       m,
	}
}

func (a *App) makeExecutionServer() *ExecutionServiceServerV1Alpha2 {
	return NewExecutionServiceServerV1Alpha2(a.messenger)
}

func (a *App) setupRestRoutes() {
	a.restRouter.HandleFunc("/block/{height}", a.getBlock).Methods("GET")
	a.restRouter.HandleFunc("/message", a.postMessage).Methods("POST")
}

func (a *App) makeRestServer() *http.Server {
	return &http.Server{
		Addr:    a.restAddr,
		Handler: a.restRouter,
	}
}

func (a *App) Run() {
	// run execution api
	go func() {
		server := a.makeExecutionServer()
		lis, err := net.Listen("tcp", a.executionAddr)
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
		fmt.Printf("rest api server closed\n")
	} else if err != nil {
		fmt.Printf("error listening for rest api server: %s\n", err)
	}

}
