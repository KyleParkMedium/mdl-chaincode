package chaincode

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	// Define key names for options
	totalSupplyKey = "totalSupply"
	tokenName      = "mdl"

	// Define objectType names for prefix
	// minterKeyPrefix = "minter"
	allowancePrefix = "allowance"

	// Define client, 자산 분할
	clientPrefix               = "client"
	clientPartitionPrefix      = "clientPartition"
	clientPartitionIndexPrefix = "clientPartitionIndex"

	// AddressLength is the expected length of the address
	addressLength = 20
)

// / @notice Use to get the list of partitions `_tokenHolder` is associated with
// / @param _tokenHolder An address corresponds whom partition list is queried
// / @return List of partitions
func (s *SmartContract) partitionsOf(ctx contractapi.TransactionContextInterface, _tokenHolder string) (string, error) {

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientPartitionPrefix, []string{_tokenHolder})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPartitionPrefix, err)
	}

	partitionsListBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}

	if partitionsListBytes == nil {
		return 0, fmt.Errorf("the account %s does not exist", _tokenHolder)
	}

	partitionsListJSON := json.Unmarshal(partitionsListBytes)
	var partitionsList []string
	for i := 0; i < partitionsListJSON.length; i++ {
		partitionsList[i] = partitions[_tokenHolder][i].partition
	}

	return partitionsList
}

func (s *SmartContract) AdminMint(ctx contractapi.TransactionContextInterface, amount int) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}

	// Get ID of submitting client identity
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}
	address := getAddress([]byte(id))

	randomByte := make([]byte, 20)
	rand.Read(randomByte)
	randomAddress := getAddress(randomByte)

	minterKey, err := ctx.GetStub().CreateCompositeKey(minterKeyPrefix, []string{address})

	token := PartitionToken{tokenName, clientMSPID, false, {amount: 200, Address: randomAddress}}
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}

	err = ctx.GetStub().PutState(minterKey, tokenJSON)

	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	log.Printf("minter account %s registered with %s", address, string(tokenJSON))

	return nil
}
