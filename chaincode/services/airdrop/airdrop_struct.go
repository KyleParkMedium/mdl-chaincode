package airdrop

import "github.com/KyleParkMedium/mdl-chaincode/chaincode/services/token"

type AirDropStruct struct {
	DocType string `json:"docType"`

	Caller string

	Array []token.MintByPartitionStruct
}
