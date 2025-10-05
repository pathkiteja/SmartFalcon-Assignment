package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var gateway *client.Gateway
var contract *client.Contract

// config paths (adjust to your test-network paths)
var (
	mspDir   = getenv("MSP_DIR", "/app/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp")
	certPath = filepath.Join(mspDir, "signcerts", "cert.pem")

	keyDir      = filepath.Join(mspDir, "keystore")
	tlsCertPath = "/app/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"

	// peer must be reachable inside Docker network
	peerEndpoint  = "peer0.org1.example.com:7051"
	gatewayPeer   = "peer0.org1.example.com"
	channelName   = "mychannel"
	chaincodeName = "accountcc"
)

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// helper to read private key path from keystore
func findKeyFile(dir string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if !f.IsDir() {
			return filepath.Join(dir, f.Name()), nil
		}
	}
	return "", fmt.Errorf("no key file in %s", dir)
}

func connectGateway() error {
	// Read the certificate PEM file
	certPEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %v", err)
	}

	// Parse certificate
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Create identity using the parsed certificate
	id, err := identity.NewX509Identity("Org1MSP", cert)
	if err != nil {
		return fmt.Errorf("failed to create identity: %v", err)
	}

	// Read private key
	keyFile, err := findKeyFile(keyDir)
	if err != nil {
		return fmt.Errorf("failed to find key file: %v", err)
	}
	privateKeyPEM, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key: %v", err)
	}

	// Create signer from private key PEM
	// Parse private key and create signer
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block for private key")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	signer, err := identity.NewPrivateKeySign(privKey)
	if err != nil {
		return fmt.Errorf("failed to create private key signer: %v", err)
	}

	// TLS credentials for gRPC
	tlsCert, err := ioutil.ReadFile(tlsCertPath)
	if err != nil {
		return fmt.Errorf("failed to read TLS certificate: %v", err)
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(tlsCert); !ok {
		return fmt.Errorf("failed to add TLS cert to pool")
	}
	creds := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	conn, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %v", err)
	}

	// Connect to gateway
	gw, err := client.Connect(
		id,
		client.WithSign(signer),
		client.WithClientConnection(conn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(30*time.Second),
		client.WithCommitStatusTimeout(30*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %v", err)
	}

	gateway = gw
	network := gateway.GetNetwork(channelName)
	contract = network.GetContract(chaincodeName)
	return nil
}

func closeGateway() {
	if gateway != nil {
		gateway.Close()
	}
}

// REST handlers

type CreateAccountRequest struct {
	Key         string `json:"key"`
	DealerID    string `json:"DEALERID"`
	MSISDN      string `json:"MSISDN"`
	MPIN        string `json:"MPIN"`
	Balance     string `json:"BALANCE"`
	Status      string `json:"STATUS"`
	TransAmount string `json:"TRANSAMOUNT"`
	TransType   string `json:"TRANSTYPE"`
	Remarks     string `json:"REMARKS"`
}

func createAccountHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := contract.SubmitTransaction("CreateAccount",
		req.Key, req.DealerID, req.MSISDN, req.MPIN, req.Balance, req.Status, req.TransAmount, req.TransType, req.Remarks)
	if err != nil {
		http.Error(w, fmt.Sprintf("SubmitTransaction failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"result":"created","key":"%s"}`, req.Key)
}

func readAccountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	eval, err := contract.EvaluateTransaction("ReadAccount", key)
	if err != nil {
		http.Error(w, fmt.Sprintf("EvaluateTransaction failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(eval)
}

type UpdateRequest struct {
	DealerID    string `json:"DEALERID,omitempty"`
	MSISDN      string `json:"MSISDN,omitempty"`
	MPIN        string `json:"MPIN,omitempty"`
	Balance     string `json:"BALANCE,omitempty"`
	Status      string `json:"STATUS,omitempty"`
	TransAmount string `json:"TRANSAMOUNT,omitempty"`
	TransType   string `json:"TRANSTYPE,omitempty"`
	Remarks     string `json:"REMARKS,omitempty"`
}

func updateAccountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := contract.SubmitTransaction("UpdateAccount", key, req.DealerID, req.MSISDN, req.MPIN, req.Balance, req.Status, req.TransAmount, req.TransType, req.Remarks)
	if err != nil {
		http.Error(w, fmt.Sprintf("SubmitTransaction failed: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, `{"result":"updated","key":"%s"}`, key)
}

func accountHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	eval, err := contract.EvaluateTransaction("GetAccountHistory", key)
	if err != nil {
		http.Error(w, fmt.Sprintf("EvaluateTransaction failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(eval)
}

func main() {
	if err := connectGateway(); err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer closeGateway()

	r := mux.NewRouter()
	r.HandleFunc("/accounts", createAccountHandler).Methods("POST")
	r.HandleFunc("/accounts/{id}", readAccountHandler).Methods("GET")
	r.HandleFunc("/accounts/{id}", updateAccountHandler).Methods("PUT")
	r.HandleFunc("/accounts/{id}/history", accountHistoryHandler).Methods("GET")

	addr := ":8080"
	fmt.Printf("REST API listening on %s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}