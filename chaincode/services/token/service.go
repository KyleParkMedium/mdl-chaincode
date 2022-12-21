package token

import (
	"encoding/json"
	"fmt"

	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ccutils"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ledgermanager"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func TotalSupply(ctx contractapi.TransactionContextInterface) (*TotalSupplyStruct, error) {

	totalSupplyBytes, err := ledgermanager.GetState(DocType_TotalSupply, "TotalSupply", ctx)
	if err != nil {
		return nil, err
	}

	totalSupply := TotalSupplyStruct{}
	if err := json.Unmarshal(totalSupplyBytes, &totalSupply); err != nil {
		return nil, err
	}
	return &totalSupply, nil
}

func TotalSupplyByPartition(ctx contractapi.TransactionContextInterface, partition string) (*TotalSupplyByPartitionStruct, error) {

	totalSupplyBytes, err := ledgermanager.GetState(DocType_TotalSupplyByPartition, partition, ctx)
	if err != nil {
		return nil, err
	}

	totalSupplyByPartition := TotalSupplyByPartitionStruct{}
	if err := json.Unmarshal(totalSupplyBytes, &totalSupplyByPartition); err != nil {
		return nil, err
	}
	return &totalSupplyByPartition, nil
}

func BalanceOfByPartition(ctx contractapi.TransactionContextInterface, _tokenHolder string, _partition string) (int64, error) {

	// Create allowanceKey
	walletKey, err := ctx.GetStub().CreateCompositeKey(BalancePrefix, []string{_tokenHolder, _partition})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", BalancePrefix, err)
	}

	// key 어떻게 할건지
	partitionTokenBytes, err := ledgermanager.GetState(DocType_Token, walletKey, ctx)
	if err != nil {
		return 0, err
	}

	partitionToken := PartitionToken{}
	if err := json.Unmarshal(partitionTokenBytes, &partitionToken); err != nil {
		return 0, err
	}
	return partitionToken.Amount, nil
}

func AllowanceByPartition(ctx contractapi.TransactionContextInterface, owner string, spender string, partition string) (*AllowanceByPartitionStruct, error) {

	// Create allowanceKey
	allowancePartitionKey, err := ctx.GetStub().CreateCompositeKey(allowanceByPartitionPrefix, []string{owner, spender, partition})
	if err != nil {
		return nil, fmt.Errorf("failed to create the composite key for prefix %s: %v", allowanceByPartitionPrefix, err)
	}

	allowanceBytes, err := ledgermanager.GetState(DocType_Allowance, allowancePartitionKey, ctx)
	if err != nil {
		return nil, err
	}

	allowanceByPartition := AllowanceByPartitionStruct{}
	if err := json.Unmarshal(allowanceBytes, &allowanceByPartition); err != nil {
		return nil, err
	}
	return &allowanceByPartition, nil
}

func ApproveByPartition(ctx contractapi.TransactionContextInterface, allowanceByPartition AllowanceByPartitionStruct) error {

	allowanceByPartitionToMap, err := ccutils.StructToMap(allowanceByPartition)

	// Create allowanceKey
	allowancePartitionKey, err := ctx.GetStub().CreateCompositeKey(allowanceByPartitionPrefix, []string{allowanceByPartition.Owner, allowanceByPartition.Spender, allowanceByPartition.Partition})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", allowanceByPartitionPrefix, err)
	}

	exist, err := ledgermanager.CheckExistState(allowancePartitionKey, ctx)
	if err != nil {
		return err
	}

	if exist {
		err = ledgermanager.UpdateState(DocType_Allowance, allowancePartitionKey, allowanceByPartitionToMap, ctx)
		if err != nil {
			return err
		}
	} else {
		_, err = ledgermanager.PutState(DocType_Allowance, allowancePartitionKey, allowanceByPartition, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func IssuanceAsset(ctx contractapi.TransactionContextInterface, token PartitionToken) (*PartitionToken, error) {

	// // key 어떻게 할건지
	// tokenBytes, err := ledgermanager.GetState(DocType_Token, token.TokenID, ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// if tokenBytes != nil {
	// 	return nil, fmt.Errorf("The asset is already registered : %s, %v", token.TokenID, err)
	// }

	// token
	_, err := ledgermanager.PutState(DocType_Token, token.TokenID, token, ctx)
	if err != nil {
		return nil, err
	}

	// totalSupplyByPartition
	_, err = ledgermanager.PutState(DocType_TotalSupplyByPartition, token.TokenID, TotalSupplyByPartitionStruct{TotalSupply: 0, Partition: token.TokenID}, ctx)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// /** 사용자 지갑 생성 */
// func (s *SmartContract) CreateWalletMB(ctx contractapi.TransactionContextInterface, args map[string]interface{}) (string, error) {

// 	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
// 	err := _getMSPID(ctx)
// 	if err != nil {
// 		return "", err
// 	}

// 	id, err := _msgSender(ctx)
// 	if err != nil {
// 		return "", err
// 	}

// 	// owner Address
// 	owner := ccutils.GetAddress([]byte(id))

// 	// Create allowanceKey
// 	clientKey, err := ctx.GetStub().CreateCompositeKey(clientBalancePrefix, []string{owner})
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create the composite key for prefix %s: %v", clientBalancePrefix, err)
// 	}

// 	clientWalletBytes, err := ctx.GetStub().GetState(clientKey)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read minter account from world state: %v", err)
// 	}
// 	if clientWalletBytes != nil {
// 		return "", fmt.Errorf("client %s has been already registered: %v", owner, err)
// 	}

// 	wallet := Wallet{Name: owner}

// 	walletJSON, err := json.Marshal(wallet)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to obtain JSON encoding: %v", err)
// 	}
// 	err = ctx.GetStub().PutState(clientKey, walletJSON)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to put state: %v", err)
// 	}

// 	log.Printf("client account %s registered", owner)

// 	return owner, nil
// }

// /** 파티션별 사용자 업데이트 */
// func (s *SmartContract) PartitionArrayUpdate(ctx contractapi.TransactionContextInterface, partition string) (string, error) {

// 	// 원장에 업데이트가 될텐데
// 	// 이들을 key로 구분?
// 	// 아니면 key로 불러와 다시 특정 구조체로 파싱하는 것까지?

// 	return "hi", nil

// }
