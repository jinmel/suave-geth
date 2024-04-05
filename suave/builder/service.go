package builder

import (
	"net/http"

	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/mux"
)

type RelayService struct {
	srv   *http.Server
	relay *LocalRelay
}

func (s *RelayService) Start() error {
	if s.srv != nil {
		log.Info("Service started")
		go s.srv.ListenAndServe()
	}

	s.relay.Start()

	return nil
}

func (s *RelayService) Stop() error {
	if s.srv != nil {
		s.srv.Close()
	}
	s.relay.Stop()
	return nil
}

func NewRelayService(listenAddr string, relay *LocalRelay) *RelayService {
	srv := &http.Server{
		Addr:    listenAddr,
		Handler: getRouter(relay),
	}

	return &RelayService{
		srv:   srv,
		relay: relay,
	}
}

func getRouter(relay *LocalRelay) http.Handler {
	handler := mux.NewRouter()
	handler.HandleFunc("/eth/v1/builder/payload/{slot:[0-9]+}/{parentHash}", relay.GetPayload).Methods("GET")
	return handler
}
