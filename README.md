⚙️ Hyperledger Fabric Account Management System

A blockchain-based account management system built on Hyperledger Fabric, providing a secure, auditable, and immutable way to manage accounts with Go-based chaincode and REST APIs.

📁 Project Structure
asset-transfer-assignment/
├── chaincode-go/           # Smart contract (chaincode)
│   ├── account_chaincode.go
│   ├── go.mod
│   └── go.sum
└── rest-api/            # REST API service
    ├── main.go
    ├── Dockerfile
    ├── go.mod
    └── go.sum

✅ Prerequisites

Go 1.23+

Docker & Docker Compose

Hyperledger Fabric 2.5.x

Fabric Samples Repository

🧩 Setup Instructions
1. Install Hyperledger Fabric

Follow the official Fabric Getting Started Guide
.

2. Clone Fabric Samples
curl -sSL https://bit.ly/2ysbOFE | bash -s
cd fabric-samples

3. Start the Test Network
cd test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca

4. Deploy the Chaincode
./network.sh deployCC -ccn accountcc -ccp ../asset-transfer-assignment/chaincode-go -ccl go

🔗 Smart Contract (Chaincode)
Account Structure
type Account struct {
    DealerID    string  `json:"DEALERID"`
    MSISDN      string  `json:"MSISDN"`
    MPIN        string  `json:"MPIN"`
    Balance     float64 `json:"BALANCE"`
    Status      string  `json:"STATUS"`
    TransAmount float64 `json:"TRANSAMOUNT"`
    TransType   string  `json:"TRANSTYPE"`
    Remarks     string  `json:"REMARKS"`
}

Chaincode Functions

CreateAccount – Create a new account

ReadAccount – Read account details

UpdateAccount – Update account information

GetAccountHistory – Retrieve transaction history

AccountExists – Check if account exists

🌐 REST API
Build Docker Image
cd asset-transfer-assignment/rest-api
docker build -t asset-rest:latest .

Run REST API
docker run --rm -p 8080:8080 \
--network fabric_test \
-v $(pwd)/../../test-network/organizations:/app/organizations:ro \
asset-rest:latest

📡 API Endpoints
🧾 Create Account
curl -X POST http://localhost:8080/accounts \
-H "Content-Type: application/json" \
-d '{
  "key": "acc001",
  "DEALERID": "D001",
  "MSISDN": "9876543210",
  "MPIN": "1234",
  "BALANCE": "10000",
  "STATUS": "ACTIVE",
  "TRANSAMOUNT": "0",
  "TRANSTYPE": "INIT",
  "REMARKS": "Initial account"
}'


Response

{"result":"created","key":"acc001"}

📖 Read Account
curl http://localhost:8080/accounts/acc001


Response

{
  "DEALERID": "D001",
  "MSISDN": "9876543210",
  "MPIN": "1234",
  "BALANCE": 10000,
  "STATUS": "ACTIVE",
  "TRANSAMOUNT": 0,
  "TRANSTYPE": "INIT",
  "REMARKS": "Initial account"
}

✏️ Update Account
curl -X PUT http://localhost:8080/accounts/acc001 \
-H "Content-Type: application/json" \
-d '{
  "BALANCE": "15000",
  "REMARKS": "Balance updated"
}'


Response

{"result":"updated","key":"acc001"}

🕓 Get Account History
curl http://localhost:8080/accounts/acc001/history


Response

[
  {
    "IsDelete": false,
    "Timestamp": {
      "seconds": 1759397987,
      "nanos": 869326844
    },
    "TxId": "5c021a04...",
    "Value": {
      "BALANCE": 15000,
      "DEALERID": "D001",
      "MPIN": "1234",
      "MSISDN": "9876543210",
      "REMARKS": "Balance updated",
      "STATUS": "ACTIVE",
      "TRANSAMOUNT": 0,
      "TRANSTYPE": "INIT"
    }
  }
]

🧪 Testing Chaincode Directly
Set Environment Variables
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

Create an Account
peer chaincode invoke -o localhost:7050 \
--ordererTLSHostnameOverride orderer.example.com \
--tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
-C mychannel -n accountcc \
--peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
--peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
-c '{"function":"CreateAccount","Args":["acc001","D001","9876543210","1234","10000","ACTIVE","0","INIT","Initial account"]}'

Query an Account
peer chaincode query -C mychannel -n accountcc \
-c '{"function":"ReadAccount","Args":["acc001"]}'

🧹 Cleanup
cd test-network
./network.sh down

🏗️ Architecture Overview
Components

Chaincode – Go-based smart contract managing account data

REST API – Go-based HTTP server interacting via Fabric Gateway

Fabric Network – Two orgs (Org1, Org2) + one orderer (Raft consensus)

Data Flow
Client (curl/Postman)
    ↓
REST API (Docker)
    ↓
Fabric Gateway SDK
    ↓
Peer Nodes (Org1, Org2)
    ↓
Chaincode (Go)
    ↓
Ledger (World State + Blockchain)

✨ Features

Account creation and validation

Secure updates and history tracking

RESTful API integration

Immutable and auditable ledger

Multi-organization endorsement

🔒 Security Considerations

TLS-enabled communication

Certificate-based authentication

Multi-org endorsement policies

MPIN hashing (recommended for production)

Secure MSP key management

🚀 Future Enhancements

Account-to-account transactions

Role-based access control

Real-time balance validation

Event notifications

Pagination for history

Enhanced logging & error handling

📜 License

This project was developed as part of a Hyperledger Fabric Internship Assignment.
