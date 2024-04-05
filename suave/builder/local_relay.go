package builder

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	builderApi "github.com/attestantio/go-builder-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	eth2UtilBellatrix "github.com/attestantio/go-eth2-client/util/bellatrix"
	"github.com/ethereum/go-ethereum/log"
)

type LocalRelay struct {
	lock    sync.Mutex
	payload *bellatrix.ExecutionPayload
	header  *bellatrix.ExecutionPayloadHeader
}

func NewLocalRelay() *LocalRelay {
	return &LocalRelay{}
}

func (r *LocalRelay) Start() error {
	return nil
}

func (r *LocalRelay) Stop() {
}

func (r *LocalRelay) SubmitPayload(payload *bellatrix.ExecutionPayload) error {
	header, err := PayloadToPayloadHeader(payload)
	if err != nil {
		log.Error("could not convert payload to header", "err", err)
		return err
	}

	r.lock.Lock()
	r.payload = payload
	r.header = header
	r.lock.Unlock()

	return nil
}

func PayloadToPayloadHeader(p *bellatrix.ExecutionPayload) (*bellatrix.ExecutionPayloadHeader, error) {
	if p == nil {
		return nil, errors.New("nil payload")
	}

	var txs []bellatrix.Transaction
	txs = append(txs, p.Transactions...)

	transactions := eth2UtilBellatrix.ExecutionPayloadTransactions{Transactions: txs}
	txroot, err := transactions.HashTreeRoot()
	if err != nil {
		return nil, err
	}

	return &bellatrix.ExecutionPayloadHeader{
		ParentHash:       p.ParentHash,
		FeeRecipient:     p.FeeRecipient,
		StateRoot:        p.StateRoot,
		ReceiptsRoot:     p.ReceiptsRoot,
		LogsBloom:        p.LogsBloom,
		PrevRandao:       p.PrevRandao,
		BlockNumber:      p.BlockNumber,
		GasLimit:         p.GasLimit,
		GasUsed:          p.GasUsed,
		Timestamp:        p.Timestamp,
		ExtraData:        p.ExtraData,
		BaseFeePerGas:    p.BaseFeePerGas,
		BlockHash:        p.BlockHash,
		TransactionsRoot: txroot,
	}, nil
}

func (r *LocalRelay) GetPayload(w http.ResponseWriter, req *http.Request) {
	r.lock.Lock()
	payload := r.payload
	header := r.header
	r.lock.Unlock()

	if payload == nil || header == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := &builderApi.VersionedExecutionPayload{
		Version:   spec.DataVersionBellatrix,
		Bellatrix: payload,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
}

type httpErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(httpErrorResp{code, message}); err != nil {
		http.Error(w, message, code)
	}
}
