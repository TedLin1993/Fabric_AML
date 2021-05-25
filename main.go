/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

func main() {
	amlContract := new(AmlContract)
	amlContract.Info.Version = "0.0.1"
	amlContract.Info.Description = "My Smart Contract"
	amlContract.Info.License = new(metadata.LicenseMetadata)
	amlContract.Info.License.Name = "Apache-2.0"
	amlContract.Info.Contact = new(metadata.ContactMetadata)
	amlContract.Info.Contact.Name = "John Doe"

	chaincode, err := contractapi.NewChaincode(amlContract)
	chaincode.Info.Title = "Documents chaincode"
	chaincode.Info.Version = "0.0.1"

	if err != nil {
		panic("Could not create chaincode from AmlContract." + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
