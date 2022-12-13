package chaincode

import (
	"encoding/json"
	"fmt"
	"log"
	"mdl-chaincode/chaincode/ccutils"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) GetMSPID(ctx contractapi.TransactionContextInterface) string {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens

	// Get ID of submitting client identity
	id, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "error"
	}
	return id
}

func (s *SmartContract) GetID(ctx contractapi.TransactionContextInterface) string {
	id, err := _msgSender(ctx)
	if err != nil {
		return "error"
	}
	return id
}

/** 사용자 지갑 생성 */
func (s *SmartContract) CreateWallet(ctx contractapi.TransactionContextInterface) (string, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return "", err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return "", err
	}

	// owner Address
	owner := ccutils.GetAddress([]byte(id))

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientWalletPrefix, []string{owner})
	if err != nil {
		return "", fmt.Errorf("failed to create the composite key for prefix %s: %v", clientWalletPrefix, err)
	}

	clientWalletBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return "", fmt.Errorf("failed to read minter account from world state: %v", err)
	}
	if clientWalletBytes != nil {
		return "", fmt.Errorf("client %s has been already registered: %v", owner, err)
	}

	wallet := Wallet{Name: owner}

	walletJSON, err := json.Marshal(wallet)
	if err != nil {
		return "", fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(clientKey, walletJSON)
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	log.Printf("client account %s registered", owner)

	return owner, nil
}

// 토큰에 사용자를 매핑하는 법은 두 가지 방법이 있을것
// map으로 받고 파싱하던가, 아니면 초장에 구조체로 받던가
func (s *SmartContract) IssuanceAssetMapArgs(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (string, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return "", err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return "", err
	}

	// owner Address
	address := ccutils.GetAddress([]byte(id))

	// Generate random bytes of assets
	partitionAddress := ccutils.GetAddress([]byte("random bytes of assets"))
	// partitionAddress := getAddress([]byte(partition))

	// minter := string(minterBytes)
	// if minter != ccutils.GetAddress([]byte(id)) {
	// 	return fmt.Errorf("client is not authorized to mint new tokens")
	// }

	// Create Asset
	minterByPartitionKey, err := ctx.GetStub().CreateCompositeKey(mintingByPartitionPrefix, []string{address, partitionAddress})
	if err != nil {
		return "", fmt.Errorf("failed to create the composite key for prefix %s: %v", mintingByPartitionPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(minterByPartitionKey)
	if err != nil {
		return "", fmt.Errorf("failed to read minter account %s from world state: %v", address, err)
	}

	if tokenBytes != nil {
		return "", fmt.Errorf("The asset is already registered : %s, %v", partitionAddress, err)
	}

	example := Partition{Amount: 0, Partition: partitionAddress}
	token := PartitionToken{Name: "tokenName", ID: address, Locked: false, Partition: example}
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(minterByPartitionKey, tokenJSON)
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	log.Printf("The Asset %s is registered", partitionAddress)

	// Create allowanceKey
	totalSupplyByPartitionKey, err := ctx.GetStub().CreateCompositeKey(totalSupplyByPartitionPrefix, []string{partitionAddress})
	if err != nil {
		return "", fmt.Errorf("failed to create the composite key for prefix %s: %v", totalSupplyByPartitionKey, err)
	}

	supplyToken := TotalSupplyByPartition{TotalSupply: 0, Partition: partitionAddress}
	supplyTokenJSON, err := json.Marshal(supplyToken)
	if err != nil {
		return "", fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(totalSupplyByPartitionKey, supplyTokenJSON)
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	return partitionAddress, nil
}

// func (s *SmartContract) AirDrop(ctx contractapi.TransactionContextInterface) error {
// 	a := ctx.GetStub().GetStateByPartialCompositeKey()
// }
