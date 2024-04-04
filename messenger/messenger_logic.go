package messenger

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Transaction represents a transaction in the blockchain.
type Transaction struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

// register rollup specific handler
func registerHandlers(a *App) {
	a.restRouter.HandleFunc("/message", a.postMessage).Methods("POST")
	a.restRouter.HandleFunc("/recent", a.getRecentMessages).Methods("GET")
}

// encode transaction into bytes to be sent to the sequencer
func encodeTx(tx Transaction) ([]byte, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		log.Errorf("error encoding transaction: %s\n", err)
		return nil, err
	}
	return data, nil
}

// decode transaction from bytes back into rollup format
func decodeTx(txEncoded []byte) (*Transaction, error) {
	tx := &Transaction{}
	if err := json.Unmarshal(txEncoded, tx); err != nil {
		log.Errorf("error decoding transaction: %s\n", err)
		return nil, err
	}
	return tx, nil
}

// create starting block for this rollup
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

// send rollup message transaction to the sequencer
func (a *App) postMessage(w http.ResponseWriter, r *http.Request) {
	var tx Transaction
	// decode transaction to ensure proper format
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		log.Errorf("error decoding transaction: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// recode transaction to send to sequencer
	txEncoded, err := encodeTx(tx)
	if err != nil {
		log.Errorf("error re-encoding transaction: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// send transaction to the sequencer
	err = a.sequencerClient.SendMessageViaComposer(txEncoded)
	if err != nil {
		log.Errorf("error sending message: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Infof("succesfully sent transaction to composer")
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
				messages = append(messages, *tx)
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

func prepareBlockForClient(blockTxs [][]byte) []byte {
	// recode messages as json for front end
	transactions := []Transaction{}
	for _, encodedTx := range blockTxs {
		decodedTx, err := decodeTx(encodedTx)
		if err != nil {
			log.Errorf("error decoding tx in block: %v, error: %s\n", encodedTx, err)
			continue
		}
		transactions = append(transactions, *decodedTx)
	}

	// only write blocks with valid transactions
	if len(transactions) == 0 {
		return []byte{}
	}

	// send transaction encoded format to frontend
	txsJson, err := json.Marshal(transactions)
	if err != nil {
		log.Errorf("Failed to marshal transactions: %v", err)
		return []byte{}
	}
	log.Debugf("marshalled txsJson: %s", txsJson)

	return txsJson
}
