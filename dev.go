package chaincode

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

// 내가 하는거
// 이 함수를 init 함수로 놓고 보자
// 일단은 이렇게 넘어가는데, identity 에 대한 부분 모듈로 빼서 매개변수 처리하는 식으로 바꿀 거임.
// 완료
func (s *SmartContract) Minter(ctx contractapi.TransactionContextInterface, amount int) error {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}

	minterBytes, err := ctx.GetStub().GetState(minterKeyPrefix)
	if err != nil {
		return fmt.Errorf("failed to read minter account from world state: %v", err)
	}

	// ? 이게 굳이 왜?
	// 이 부분은 수정이 필요함. key + id 식으로 개별 정보를 확인할 수 있게 바꿔야함
	// 요딴 걸로
	// clientKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{address})
	// if minterBytes != nil {
	// 	return fmt.Errorf("minter has been already registered: %v", err)
	// }

	// Get ID of submitting client identity
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	// value, found, err := ctx.GetClientIdentity().GetAttributeValue("roles")

	address := getAddress([]byte(id))

	err = ctx.GetStub().PutState(minterKeyPrefix, []byte(address))
	// // 사실 이렇게 되야함
	// err = ctx.GetStub().PutState(id, []byte(address))
	// // 더 정확하게는
	// err = ctx.GetStub().PutState(minterKey[id], []byte(address))

	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{address})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPrefix, err)
	}

	clientAmountBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return fmt.Errorf("failed to read minter account from world state: %v", err)
	}

	if clientAmountBytes != nil {
		return fmt.Errorf("client %s has been already registered: %v", address, err)
	}

	// // init 함수에서 mint 실행 굳이 넣어야하나?
	// if amount < 0 {
	// 	return fmt.Errorf("mint amount must be a positive integer")
	// }

	// token := Token{tokenName, clientMSPID, amount, false}
	// tokenJSON, err := json.Marshal(token)
	// if err != nil {
	// 	return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	// }
	// err = ctx.GetStub().PutState(clientKey, tokenJSON)
	// if err != nil {
	// 	return fmt.Errorf("failed to put state: %v", err)
	// }

	// log.Printf("minter account %s registered with %s", address, string(tokenJSON))

	return nil
}

// func AddMinter()          // event MinterAdded
// func RenounceMinter()     // event MinterRemoved
// func TransferMinterRole() // event MinterAdded & MinterRemoved

// Mint creates new tokens and adds them to minter's account balance
// This function triggers a Transfer event
// 얘도 마찬가지로 매개변수로 빼야함 id
// 완료
func (s *SmartContract) Mint(ctx contractapi.TransactionContextInterface, amount int) error {

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

	// minterKey, err := ctx.GetStub().CreateCompositeKey(minterKeyPrefix, []string{address})

	// // 여기를 minter - 주소 매칭이 아니라, id - 주소 매칭으로 해야 될 거 같은데?
	// // // 요런 식으로
	// // minterBytes, err := ctx.GetStub().GetState(minterKey[id])
	// // minterBytes, err := ctx.GetStub().GetState(minterKey2)

	// if err != nil {
	// 	return fmt.Errorf("failed to read minter account from world state: %v", err)
	// }

	// // 여기 부분부터 해서 주석 처리하고 수정해야함
	// minter := string(minterBytes)
	// if minter != getAddress([]byte(id)) {
	// 	return fmt.Errorf("client is not authorized to mint new tokens")
	// }

	if amount <= 0 {
		return fmt.Errorf("mint amount must be a positive integer")
	}

	// Create allowanceKey
	clientKey, err := ctx.GetStub().CreateCompositeKey(clientPrefix, []string{address})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", clientPrefix, err)
	}

	tokenBytes, err := ctx.GetStub().GetState(clientKey)
	if err != nil {
		return fmt.Errorf("failed to read minter account %s from world state: %v", minter, err)
	}

	token := new(Token)
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON decoding: %v", err)
	}

	if clientMSPID != token.MSPID {
		return fmt.Errorf("client is not authorized to mint new tokens")
	}

	currentBalance := token.Amount

	updatedBalance := currentBalance + amount

	token.Amount = updatedBalance

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().PutState(clientKey, tokenJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	// Update the totalSupply
	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	var totalSupply int

	// If no tokens have been minted, initialize the totalSupply
	if totalSupplyBytes == nil {
		totalSupply = 0
	} else {
		totalSupply, _ = strconv.Atoi(string(totalSupplyBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
	}

	// Add the mint amount to the total supply and update the state
	totalSupply += amount
	err = ctx.GetStub().PutState(totalSupplyKey, []byte(strconv.Itoa(totalSupply)))
	if err != nil {
		return err
	}

	// Emit the Transfer event
	transferEvent := Event{"0x0", minter, amount}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("minter account %s balance updated from %s to %s", minter, string(tokenBytes), string(tokenJSON))

	return nil
}
