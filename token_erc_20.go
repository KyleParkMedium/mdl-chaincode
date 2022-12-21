/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	// "mdl-chaincode/chaincode"

	"github.com/KyleParkMedium/mdl-chaincode/chaincode/controller"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	// A *chaincode.SmartContract
	// chaincode.SmartContract
	controller.SmartContract
	// B *controller.SmartContract
	// contractapi.Contract
	// SmartContract2 chaincode.SmartContract
	// SmartContract2 controller.SmartContract
}

// func StructToMap(arg interface{}) (map[string]interface{}, error) {
// 	data, err := json.Marshal(arg) // Convert to a json string
// 	if err != nil {
// 		return nil, nil
// 	}

// 	result := make(map[string]interface{})
// 	err = json.Unmarshal(data, &result) // Convert to a map
// 	return result, nil
// }

func main() {

	tokenChaincode, err := contractapi.NewChaincode(&SmartContract{})

	if err != nil {
		log.Panicf("Error creating token-erc-20 chaincode: %v", err)
	}

	if err := tokenChaincode.Start(); err != nil {
		log.Panicf("Error starting token-erc-20 chaincode: %v", err)
	}
}
