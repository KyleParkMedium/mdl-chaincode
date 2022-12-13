package chaincode

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"mdl-chaincode/chaincode/ccutils"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) TestIssuanceAsset(ctx contractapi.TransactionContextInterface) (string, error) {

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
	randomByte := make([]byte, 20)
	rand.Read(randomByte)
	randomAddress := ccutils.GetAddress(randomByte)
	exampleAsset := ccutils.GetAddress([]byte("random bytes of assets"))

	// minter := string(minterBytes)
	// if minter != ccutils.GetAddress([]byte(id)) {
	// 	return fmt.Errorf("client is not authorized to mint new tokens")
	// }

	// Create Asset
	mintingByPartitionKey, err := ctx.GetStub().CreateCompositeKey(mintingByPartitionPrefix, []string{address, exampleAsset})
	if err != nil {
		return "", fmt.Errorf("failed to create the composite key for prefix %s: %v", mintingByPartitionPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(mintingByPartitionKey)
	if err != nil {
		return "", fmt.Errorf("failed to read minter account %s from world state: %v", address, err)
	}

	if tokenBytes != nil {
		return "", fmt.Errorf("The asset is already registered : %s, %v", randomAddress, err)
	}

	example := Partition{Amount: 0, Partition: exampleAsset}
	token := PartitionToken{Name: "tokenName", ID: address, Locked: false, Partition: example}
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(mintingByPartitionKey, tokenJSON)
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	log.Printf("The Asset %s is registered", randomAddress)

	// Create allowanceKey
	totalSupplyByPartitionKey, err := ctx.GetStub().CreateCompositeKey(totalSupplyByPartitionPrefix, []string{exampleAsset})
	if err != nil {
		return "", fmt.Errorf("failed to create the composite key for prefix %s: %v", totalSupplyByPartitionKey, err)
	}

	supplyToken := TotalSupplyByPartition{TotalSupply: 0, Partition: exampleAsset}
	supplyTokenJSON, err := json.Marshal(supplyToken)
	if err != nil {
		return "", fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(totalSupplyByPartitionKey, supplyTokenJSON)
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	return exampleAsset, nil
}

func (s *SmartContract) TestGetSignedProposal(ctx contractapi.TransactionContextInterface) (int, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return 0, err
	}

	// Retrieve total supply of tokens from state of smart contract
	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	var totalSupply int

	// If no tokens have been minted, return 0
	if totalSupplyBytes == nil {
		totalSupply = 0
	} else {
		totalSupply, _ = strconv.Atoi(string(totalSupplyBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
	}

	log.Printf("TotalSupply: %d tokens", totalSupply)

	return totalSupply, nil
}
