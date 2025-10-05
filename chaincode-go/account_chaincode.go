package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Account represents the asset
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

// SmartContract provides functions for managing accounts
type SmartContract struct {
	contractapi.Contract
}

// CreateAccount creates a new account asset
func (s *SmartContract) CreateAccount(ctx contractapi.TransactionContextInterface, key string, dealerID, msisdn, mpin, balanceStr, status, transAmountStr, transType, remarks string) error {
	exists, err := s.AccountExists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("account %s already exists", key)
	}

	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return fmt.Errorf("invalid balance: %v", err)
	}
	transAmount, err := strconv.ParseFloat(transAmountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid transAmount: %v", err)
	}

	account := Account{
		DealerID:    dealerID,
		MSISDN:      msisdn,
		MPIN:        mpin,
		Balance:     balance,
		Status:      status,
		TransAmount: transAmount,
		TransType:   transType,
		Remarks:     remarks,
	}

	accountBytes, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(key, accountBytes)
}

// ReadAccount returns the account stored in the world state
func (s *SmartContract) ReadAccount(ctx contractapi.TransactionContextInterface, key string) (*Account, error) {
	data, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if data == nil {
		return nil, fmt.Errorf("account %s does not exist", key)
	}

	var account Account
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

// UpdateAccount updates fields of an account.
// For simplicity, pass empty strings for fields you don't want to change.
func (s *SmartContract) UpdateAccount(ctx contractapi.TransactionContextInterface, key string, dealerID, msisdn, mpin, balanceStr, status, transAmountStr, transType, remarks string) error {
	account, err := s.ReadAccount(ctx, key)
	if err != nil {
		return err
	}

	if dealerID != "" {
		account.DealerID = dealerID
	}
	if msisdn != "" {
		account.MSISDN = msisdn
	}
	if mpin != "" {
		account.MPIN = mpin
	}
	if balanceStr != "" {
		b, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return fmt.Errorf("invalid balance: %v", err)
		}
		account.Balance = b
	}
	if status != "" {
		account.Status = status
	}
	if transAmountStr != "" {
		ta, err := strconv.ParseFloat(transAmountStr, 64)
		if err != nil {
			return fmt.Errorf("invalid transAmount: %v", err)
		}
		account.TransAmount = ta
	}
	if transType != "" {
		account.TransType = transType
	}
	if remarks != "" {
		account.Remarks = remarks
	}

	bytes, err := json.Marshal(account)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(key, bytes)
}

// GetAccountHistory returns the history for a key
func (s *SmartContract) GetAccountHistory(ctx contractapi.TransactionContextInterface, key string) ([]map[string]interface{}, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []map[string]interface{}
	for resultsIterator.HasNext() {
		h, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var value interface{}
		if len(h.Value) > 0 {
			if err := json.Unmarshal(h.Value, &value); err != nil {
				// in some cases an empty value (delete) may be returned
				value = string(h.Value)
			}
		}
		record := map[string]interface{}{
			"TxId":      h.TxId,
			"Timestamp": h.Timestamp, // protobuf Timestamp; JSON marshalling may need formatting in client
			"IsDelete":  h.IsDelete,
			"Value":     value,
		}
		records = append(records, record)
	}
	return records, nil
}

// AccountExists returns true when account with given key exists in world state
func (s *SmartContract) AccountExists(ctx contractapi.TransactionContextInterface, key string) (bool, error) {
	data, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, err
	}
	return data != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error create account chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting account chaincode: %s", err.Error())
	}
}