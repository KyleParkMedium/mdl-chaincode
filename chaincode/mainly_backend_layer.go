package chaincode

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ccutils"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) IssuanceAssetMB(ctx contractapi.TransactionContextInterface, partition string, aa ...int) (string, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	err := _getMSPID(ctx)
	if err != nil {
		return "", err
	}

	id, err := _msgSender(ctx)
	if err != nil {
		return "", err
	}

	// check operator?? or role

	// owner Address
	address := ccutils.GetAddress([]byte(id))

	// Generate random bytes of assets
	partitionAddress := ccutils.GetAddress([]byte(partition))

	// Create Asset
	mintingByPartitionKey, err := ctx.GetStub().CreateCompositeKey(mintingByPartitionPrefix, []string{address, partitionAddress})
	if err != nil {
		return "", fmt.Errorf("failed to create the composite key for prefix %s: %v", mintingByPartitionPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(mintingByPartitionKey)
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
	err = ctx.GetStub().PutState(mintingByPartitionKey, tokenJSON)
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

/** 사용자 지갑 생성 */
func (s *SmartContract) CreateWalletMB(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (string, error) {

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

/** 파티션별 사용자 업데이트 */
func (s *SmartContract) PartitionArrayUpdate(ctx contractapi.TransactionContextInterface, partition string) (string, error) {

	// 원장에 업데이트가 될텐데
	// 이들을 key로 구분?
	// 아니면 key로 불러와 다시 특정 구조체로 파싱하는 것까지?

	return "hi", nil

}
