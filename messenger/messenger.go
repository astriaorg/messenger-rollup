package messenger

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	astriaGrpc "buf.build/gen/go/astria/execution-apis/grpc/go/astria/execution/v1alpha2/executionv1alpha2grpc"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

type Config struct {
	SequencerRPC string `env:"SEQUENCER_RPC, required"`
	ConductorRPC string `env:"CONDUCTOR_RPC, required"`
	RESTApiPort  string `env:"RESTAPI_PORT, required"`
	RollupName   string `env:"ROLLUP_NAME, required"`
	SeqPrivate   string `env:"SEQUENCER_PRIVATE, required"`
}

// Transaction represents a transaction in the blockchain.
type Transaction struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

// App is the main application struct, containing all the necessary components.
type App struct {
	executionRPC    string
	sequencerRPC    string
	sequencerClient SequencerClient
	restRouter      *mux.Router
	restAddr        string
	rollupBlocks    *RollupBlocks
	rollupName      string
	rollupID        []byte
	newBlockChan    chan Block
	wsClients       WSClientList
	sync.RWMutex
}

func encodeTx(tx Transaction) ([]byte, error) {
	// encode transaction to send to sequencer
	data, err := json.Marshal(tx)
	if err != nil {
		log.Errorf("error encoding transaction: %s\n", err)
		return data, err
	}
	return data, nil
}

func decodeTx(txEncoded []byte) (Transaction, error) {
	tx := &Transaction{}
	if err := json.Unmarshal(txEncoded, tx); err != nil {
		log.Errorf("error decoding transaction: %s\n", err)
		return *tx, err
	}
	return *tx, nil
}

func GenesisTransaction() []byte {
	genesisTx := Transaction{
		Sender:  "astria",
		Message: "hello, world!",
	}
	encodedTx, err := encodeTx(genesisTx)
	if err != nil {
		log.Errorf("error encoding genesis tx: %s\n", err)
		panic(err)
	}
	return encodedTx
}

func NewApp(cfg Config) *App {
	log.Debugf("Creating new messenger app with config: %v", cfg)

	newBlockChan := make(chan Block, 20)
	rollupBlocks := NewRollupBlocks(newBlockChan)
	router := mux.NewRouter()

	rollupID := sha256.Sum256([]byte(cfg.RollupName))

	// sequencer private key
	privateKeyBytes, err := hex.DecodeString(cfg.SeqPrivate)
	if err != nil {
		panic(err)
	}
	private := ed25519.NewKeyFromSeed(privateKeyBytes)

	return &App{
		executionRPC:    cfg.ConductorRPC,
		sequencerRPC:    cfg.SequencerRPC,
		sequencerClient: *NewSequencerClient(cfg.SequencerRPC, rollupID[:], private),
		restRouter:      router,
		restAddr:        cfg.RESTApiPort,
		rollupBlocks:    rollupBlocks,
		rollupName:      cfg.RollupName,
		rollupID:        rollupID[:],
		newBlockChan:    newBlockChan,
		wsClients:       make(WSClientList),
	}
}

// makeExecutionServer creates a new ExecutionServiceServer.
func (a *App) makeExecutionServer() *ExecutionServiceServerV1Alpha2 {
	return NewExecutionServiceServerV1Alpha2(a.rollupBlocks, a.rollupID)
}

// setupRestRoutes sets up the routes for the REST API.
func (a *App) setupRestRoutes() {
	a.restRouter.HandleFunc("/block/{height}", a.getBlock).Methods("GET")
	a.restRouter.HandleFunc("/message", a.postMessage).Methods("POST")
	a.restRouter.HandleFunc("/recent", a.getRecentMessages).Methods("GET")
	a.restRouter.HandleFunc("/ws", a.serveWS)
}

// makeRestServer creates a new HTTP server for the REST API.
func (a *App) makeRestServer() *http.Server {
	return &http.Server{
		Addr:    a.restAddr,
		Handler: cors.Default().Handler(a.restRouter),
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
	block, err := a.rollupBlocks.GetSingleBlock(uint32(height))
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

func (a *App) getRecentMessages(w http.ResponseWriter, _ *http.Request) {
	var messages []Transaction
	for i := len(a.rollupBlocks.Blocks); i > 0; i-- {
		block := a.rollupBlocks.Blocks[i-1]
		if len(block.Txs) > 0 {
			log.Infof("block txs: %v", block.Txs)
			for _, txRaw := range block.Txs {
				log.Infof("raw txn: %v", txRaw)
				tx, err := decodeTx(txRaw)
				if err != nil {
					log.Errorf("tried to decode malformed message: %s\n", err)
					continue
				}
				messages = append(messages, tx)
			}
			if len(messages) >= 100 {
				break
			}
		}
	}

	// keep only the most recent 100
	if len(messages) > 100 {
		messages = messages[len(messages)-100:]
	}

	messagesJson, err := json.Marshal(messages)
	if err != nil {
		log.Errorf("error marshalling messages: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(messagesJson)
}

func (a *App) postMessage(w http.ResponseWriter, r *http.Request) {
	var tx Transaction
	// decode transaction to ensure proper format
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		log.Errorf("error decoding transaction: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Infof("'postMessage': tx decoded successfully: %v, ", tx)
	// recode transaction to send to sequencer
	txEncoded, err := encodeTx(tx)
	if err != nil {
		log.Errorf("error re-encoding transaction: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resp, err := a.sequencerClient.SendMessage(txEncoded)
	if err != nil {
		log.Errorf("error sending message: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.WithField("responseCode", resp.Code).Debug("transaction submission result")
}

func (a *App) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Failed to upgrade HTTP to WebSocket: %v", err)
		return
	}

	client := NewWSClient(conn, a)
	a.addWSClient(client)
	go client.WaitForMessages()
	log.Debug("new ws client connected")
}

func (a *App) addWSClient(client *WSClient) {
	a.Lock()
	defer a.Unlock()
	a.wsClients[client] = true
}

func (a *App) removeWSClient(client *WSClient) {
	a.Lock()
	defer a.Unlock()
	if _, ok := a.wsClients[client]; ok {
		// close connection
		client.conn.Close()
		// remove
		delete(a.wsClients, client)
	}
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

	log.Infof("API server listening on %s\n", a.restAddr)
	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Warnf("rest api server closed\n")
		} else if err != nil {
			log.Errorf("error listening for rest api server: %s\n", err)
		}
	}()

	// send new messages to all connected ws clients
	go func() {
		for block := range a.newBlockChan {
			// only write blocks with transactions
			if len(block.Txs) == 0 {
				continue
			}
			// re-encode messages as json for front end
			transactions := []Transaction{}
			for _, encodedTx := range block.Txs {
				decodedTx, err := decodeTx(encodedTx)
				if err != nil {
					log.Errorf("error decoding tx in block: %v, error: %s\n", encodedTx, err)
					continue
				}
				transactions = append(transactions, decodedTx)
			}

			// only write blocks with valid transactions
			if len(transactions) == 0 {
				continue
			}

			// send transaction encoded format to frontend
			txsJson, err := json.Marshal(transactions)
			log.Debugf("marshalled txsJson: %s", txsJson)
			if err != nil {
				log.Errorf("Failed to marshal transactions: %v", err)
			} else {
				for client := range a.wsClients {
					select {
					case client.egress <- txsJson:
					default:
						log.Warnf("Could not send transactions to ws client: %s", txsJson)
					}
				}
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info("Shutting down server...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
	log.Info("Server gracefully stopped")
}
