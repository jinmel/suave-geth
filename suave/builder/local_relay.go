package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	builderDeneb "github.com/attestantio/go-builder-client/api/deneb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/mux"
)

type LocalRelay struct {
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
	log.Info("SubmitBlock payload", "slot", payload.ExecutionPayload.BlockNumber, "hash", payload.ExecutionPayload.BlockHash)

	if r.payload != nil && r.payload.ExecutionPayload.BlockNumber > payload.ExecutionPayload.BlockNumber {
		log.Error("Payload already exists or slot is lower than current slot", "currentSlot", r.payload.ExecutionPayload.BlockNumber, "requestedSlot", payload.ExecutionPayload.BlockNumber)
		return fmt.Errorf("payload already exists")
	}

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

	log.Info("GetPayload request received slot: %d parentHash: %s\n", slot, parentHash.String())

	payload := r.payload

	if parentHash != common.Hash(payload.ExecutionPayload.ParentHash) || slot != payload.ExecutionPayload.BlockNumber {
		log.Error("Requested Payload does not exist", "slot", slot, "parentHash", parentHash.String())
		respondError(w, http.StatusNotFound, fmt.Sprintf("payload not found for slot %d and parent hash %s current slot %d current hash %s", slot, parentHash.String(), payload.ExecutionPayload.BlockNumber, payload.ExecutionPayload.ParentHash.String()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

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
