package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	builderDeneb "github.com/attestantio/go-builder-client/api/deneb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/mux"
)

type LocalRelay struct {
	lock    sync.Mutex
	payload *builderDeneb.SubmitBlockRequest
}

func NewLocalRelay() *LocalRelay {
	return &LocalRelay{}
}

func (r *LocalRelay) Start() error {
	return nil
}

func (r *LocalRelay) Stop() {
}

func (r *LocalRelay) SubmitBlock(ctx context.Context, payload *builderDeneb.SubmitBlockRequest) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.payload = payload

	return nil
}

func (r *LocalRelay) GetPayload(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	slot, err := strconv.ParseUint(vars["slot"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid slot")
		return
	}

	parentHash := common.HexToHash(vars["parentHash"])

	log.Info("GetPayload request received", "slot", slot, "parentHash", parentHash.String())

	r.lock.Lock()
	payload := r.payload
	r.lock.Unlock()

	if payload == nil {
		log.Info("Payload not ready", "slot", slot, "parentHash", parentHash.String())
		respondError(w, http.StatusNotFound, "payload not found")
		return
	}

	if slot != payload.ExecutionPayload.BlockNumber || parentHash != common.Hash(payload.ExecutionPayload.ParentHash) {
		log.Info("Payload not found", "slot", slot, "parentHash", parentHash.String())
		respondError(w, http.StatusNotFound, fmt.Sprintf("payload not found for slot %d and parent hash %s", slot, parentHash.String()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	log.Info("Sending payload", "slot", slot, "parentHash", parentHash.String(), "payload", payload)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Info("Failed to encode payload", "slot", slot, "parentHash", parentHash.String())
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
