package wallet

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ccutils"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/ledgermanager"
	"github.com/KyleParkMedium/mdl-chaincode/chaincode/services/token"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func CreateWallet(ctx contractapi.TransactionContextInterface, tokenWallet TokenWallet) (*TokenWallet, error) {
	_, err := ledgermanager.PutState(DocType_TokenWallet, tokenWallet.TokenWalletId, tokenWallet, ctx)
	if err != nil {
		return nil, err
	}
	return &tokenWallet, nil
}

func TransferByPartition(ctx contractapi.TransactionContextInterface, transferByPartition token.TransferByPartitionStruct) error {

	fromBytes, err := ledgermanager.GetState(DocType_TokenWallet, transferByPartition.From, ctx)
	if err != nil {
		return err
	}

	fromWallet := TokenWallet{}
	err = json.Unmarshal(fromBytes, &fromWallet)
	if err != nil {
		return err
	}

	if reflect.ValueOf(fromWallet.PartitionTokens[transferByPartition.Partition]).IsZero() {
		return fmt.Errorf("partition data in From Wallet does not exist")
	}

	toBytes, err := ledgermanager.GetState(DocType_TokenWallet, transferByPartition.To, ctx)
	if err != nil {
		return err
	}

	toWallet := TokenWallet{}
	err = json.Unmarshal(toBytes, &toWallet)
	if err != nil {
		return err
	}

	// 이건 기획팀과 회의가 필요함 2차 거래에서의 조건임
	// if reflect.ValueOf(toWallet.PartitionTokens[transferByPartition.Partition]).IsZero() {
	// 	return fmt.Errorf("partition data is not exist")
	// }

	fromCurrentBalance := fromWallet.PartitionTokens[transferByPartition.Partition][0].Amount
	if fromCurrentBalance < transferByPartition.Amount {
		return fmt.Errorf("client account %s has insufficient funds", fromWallet.TokenWalletId)
		// return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	// Math
	fromUpdatedBalance := fromCurrentBalance - transferByPartition.Amount
	fromWallet.PartitionTokens[transferByPartition.Partition][0].Amount = fromUpdatedBalance

	// toUpdatedBalance := toCurrentBalance + transferByPartition.Amount
	// toWallet.PartitionTokens[transferByPartition.Partition][0].Amount = toUpdatedBalance

	// 우선 이렇게 처리
	if reflect.ValueOf(toWallet.PartitionTokens[transferByPartition.Partition]).IsZero() {
		check := make(map[string][]token.PartitionToken)
		check[transferByPartition.Partition] = append(check[transferByPartition.Partition], token.PartitionToken{Amount: transferByPartition.Amount})
		toWallet.PartitionTokens = check
	} else {
		toWallet.PartitionTokens[transferByPartition.Partition][0].Amount += transferByPartition.Amount
	}

	fromToMap, err := ccutils.StructToMap(fromWallet)
	toToMap, err := ccutils.StructToMap(toWallet)

	err = ledgermanager.UpdateState(DocType_TokenWallet, transferByPartition.From, fromToMap, ctx)
	if err != nil {
		return err
	}
	err = ledgermanager.UpdateState(DocType_TokenWallet, transferByPartition.To, toToMap, ctx)
	if err != nil {
		return err
	}

	return nil
}

func MintByPartition(ctx contractapi.TransactionContextInterface, mintByPartition token.MintByPartitionStruct) error {

	// key 어떻게 할건지
	walletBytes, err := ledgermanager.GetState(DocType_TokenWallet, mintByPartition.Minter, ctx)
	if err != nil {
		return err
	}

	wallet := TokenWallet{}
	err = json.Unmarshal(walletBytes, &wallet)
	if err != nil {
		return err
	}

	exist := true

	if reflect.ValueOf(wallet.PartitionTokens[mintByPartition.Partition]).IsZero() {
		exist = false
	}

	// wallet.PartitionTokens[mintByPartition.Partition] = append(wallet.PartitionTokens[mintByPartition.Partition], token.PartitionToken{Amount: mintByPartition.Amount})
	// wallet.PartitionTokens[mintByPartition.Partition] = append(wallet.PartitionTokens[mintByPartition.Partition], token.PartitionToken{Amount: mintByPartition.Amount})

	// wallet.Par
	// BalaneOf
	var afterBalance int64
	balanceKey, err := ctx.GetStub().CreateCompositeKey(token.BalancePrefix, []string{mintByPartition.Minter, mintByPartition.Partition})

	if exist {
		wallet.PartitionTokens[mintByPartition.Partition][0].Amount += mintByPartition.Amount

		mintByPartitionToMap, err := ccutils.StructToMap(wallet)

		err = ledgermanager.UpdateState(DocType_TokenWallet, mintByPartition.Minter, mintByPartitionToMap, ctx)
		if err != nil {
			return err
		}

		afterBalance = wallet.PartitionTokens[mintByPartition.Partition][0].Amount
		partitionToken := token.PartitionToken{Amount: afterBalance}
		balanceOfByPartitionToMap, err := ccutils.StructToMap(partitionToken)
		err = ledgermanager.UpdateState(token.DocType_Token, balanceKey, balanceOfByPartitionToMap, ctx)
		if err != nil {
			return err
		}

	} else {
		check := make(map[string][]token.PartitionToken)
		check[mintByPartition.Partition] = append(check[mintByPartition.Partition], token.PartitionToken{Amount: mintByPartition.Amount})

		wallet.PartitionTokens[mintByPartition.Partition] = check[mintByPartition.Partition]

		wallet.PartitionTokens[mintByPartition.Partition][0] = token.PartitionToken{Amount: mintByPartition.Amount}

		mintByPartitionToMap, err := ccutils.StructToMap(wallet)

		err = ledgermanager.UpdateState(DocType_TokenWallet, mintByPartition.Minter, mintByPartitionToMap, ctx)
		if err != nil {
			return err
		}

		afterBalance = wallet.PartitionTokens[mintByPartition.Partition][0].Amount
		partitionToken := token.PartitionToken{}
		partitionToken.Amount = afterBalance
		partitionToken.DocType = token.DocType_Token
		bytes, err := json.Marshal(partitionToken)
		err = ctx.GetStub().PutState(balanceKey, bytes)
		if err != nil {
			return err
		}
		// _, err = ledgermanager.PutState(token.DocType_Token, "check", partitionToken, ctx)
		// if err != nil {
		// 	return err
		// }
	}

	// Update the totalSupply, totalSupplyByPartition
	totalSupplyBytes, err := ledgermanager.GetState(token.DocType_TotalSupply, "TotalSupply", ctx)
	if err != nil {
		return err
	}

	totalSupply := token.TotalSupplyStruct{}
	if err := json.Unmarshal(totalSupplyBytes, &totalSupply); err != nil {
		return err
	}

	totalSupply.TotalSupply += mintByPartition.Amount

	totalSupplyMap, err := ccutils.StructToMap(totalSupply)

	err = ledgermanager.UpdateState(token.DocType_TotalSupply, "TotalSupply", totalSupplyMap, ctx)
	if err != nil {
		return err
	}

	// key 어떻게 할건지
	totalSupplyByPartitionBytes, err := ledgermanager.GetState(token.DocType_TotalSupplyByPartition, mintByPartition.Partition, ctx)
	if err != nil {
		return err
	}

	totalSupplyByPartition := token.TotalSupplyByPartitionStruct{}
	if err := json.Unmarshal(totalSupplyByPartitionBytes, &totalSupplyByPartition); err != nil {
		return err
	}

	totalSupplyByPartition.TotalSupply += mintByPartition.Amount

	totalSupplyByPartitionMap, err := ccutils.StructToMap(totalSupplyByPartition)

	err = ledgermanager.UpdateState(token.DocType_TotalSupplyByPartition, mintByPartition.Partition, totalSupplyByPartitionMap, ctx)
	if err != nil {
		return err
	}

	// // // Emit the Transfer event
	// // transferEvent := Event{"0x0", address, amount}
	// // transferEventJSON, err := json.Marshal(transferEvent)
	// // if err != nil {
	// // 	return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	// // }
	// // err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	// // if err != nil {
	// // 	return fmt.Errorf("failed to set event: %v", err)
	// // }

	// // log.Printf("minter account %s balance updated from %s to %s", address, string(tokenBytes), string(tokenJSON))

	return nil
}

func BurnByPartition(ctx contractapi.TransactionContextInterface, mintByPartition token.MintByPartitionStruct) error {

	// key 어떻게 할건지
	walletBytes, err := ledgermanager.GetState(DocType_TokenWallet, mintByPartition.Minter, ctx)
	if err != nil {
		return err
	}

	wallet := TokenWallet{}

	err = json.Unmarshal(walletBytes, &wallet)
	if err != nil {
		return err
	}

	if reflect.ValueOf(wallet.PartitionTokens[mintByPartition.Partition]).IsZero() {
		return fmt.Errorf("partition data is not exist")
	}

	if wallet.PartitionTokens[mintByPartition.Partition][0].Amount < mintByPartition.Amount {
		return fmt.Errorf("currentBalance is lower than input amount")
	}

	wallet.PartitionTokens[mintByPartition.Partition][0].Amount -= mintByPartition.Amount

	mintByPartitionToMap, err := ccutils.StructToMap(wallet)

	err = ledgermanager.UpdateState(DocType_TokenWallet, mintByPartition.Minter, mintByPartitionToMap, ctx)
	if err != nil {
		return err
	}

	// Update the totalSupply, totalSupplyByPartition
	totalSupplyBytes, err := ledgermanager.GetState(token.DocType_TotalSupply, "TotalSupply", ctx)
	if err != nil {
		return err
	}

	totalSupply := token.TotalSupplyStruct{}
	if err := json.Unmarshal(totalSupplyBytes, &totalSupply); err != nil {
		return err
	}

	totalSupply.TotalSupply -= mintByPartition.Amount

	totalSupplyMap, err := ccutils.StructToMap(totalSupply)

	err = ledgermanager.UpdateState(token.DocType_TotalSupply, "TotalSupply", totalSupplyMap, ctx)
	if err != nil {
		return err
	}

	// key 어떻게 할건지
	totalSupplyByPartitionBytes, err := ledgermanager.GetState(token.DocType_TotalSupplyByPartition, mintByPartition.Partition, ctx)
	if err != nil {
		return err
	}

	totalSupplyByPartition := token.TotalSupplyByPartitionStruct{}
	if err := json.Unmarshal(totalSupplyByPartitionBytes, &totalSupplyByPartition); err != nil {
		return err
	}

	totalSupplyByPartition.TotalSupply -= mintByPartition.Amount

	totalSupplyByPartitionMap, err := ccutils.StructToMap(totalSupplyByPartition)

	err = ledgermanager.UpdateState(token.DocType_TotalSupplyByPartition, mintByPartition.Partition, totalSupplyByPartitionMap, ctx)
	if err != nil {
		return err
	}

	// // // Emit the Transfer event
	// // transferEvent := Event{"0x0", address, amount}
	// // transferEventJSON, err := json.Marshal(transferEvent)
	// // if err != nil {
	// // 	return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	// // }
	// // err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	// // if err != nil {
	// // 	return fmt.Errorf("failed to set event: %v", err)
	// // }

	// // log.Printf("minter account %s balance updated from %s to %s", address, string(tokenBytes), string(tokenJSON))

	return nil
}
