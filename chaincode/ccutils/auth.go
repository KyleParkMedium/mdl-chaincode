package ccutils

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func GetID(ctx contractapi.TransactionContextInterface) (string, error) {
	// Get ID of submitting client identity
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to get client id: %v", err)
	}

	return id, nil
}

func GetMSPID(ctx contractapi.TransactionContextInterface) error {
	// Get ID of submitting client identity
	_, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	return nil
}
