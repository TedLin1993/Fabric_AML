/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AmlContract contract for managing CRUD for Aml
type AmlContract struct {
	contractapi.Contract
}

// AmlExists returns true when asset with given ID exists in world state
func (c *AmlContract) AmlExists(ctx contractapi.TransactionContextInterface, amlID string) (bool, error) {
	data, err := ctx.GetStub().GetState(amlID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateAml creates a new instance of Aml
func (c *AmlContract) CreateAml(ctx contractapi.TransactionContextInterface, amlID string, value string) error {
	exists, err := c.AmlExists(ctx, amlID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if exists {
		return fmt.Errorf("The asset %s already exists", amlID)
	}

	aml := new(Aml)
	aml.Value = value

	bytes, _ := json.Marshal(aml)

	return ctx.GetStub().PutState(amlID, bytes)
}

// ReadAml retrieves an instance of Aml from the world state
func (c *AmlContract) ReadAml(ctx contractapi.TransactionContextInterface, amlID string) (*Aml, error) {
	exists, err := c.AmlExists(ctx, amlID)
	if err != nil {
		return nil, fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("The asset %s does not exist", amlID)
	}

	bytes, _ := ctx.GetStub().GetState(amlID)

	aml := new(Aml)

	err = json.Unmarshal(bytes, aml)

	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal world state data to type Aml")
	}

	return aml, nil
}

// UpdateAml retrieves an instance of Aml from the world state and updates its value
func (c *AmlContract) UpdateAml(ctx contractapi.TransactionContextInterface, amlID string, newValue string) error {
	exists, err := c.AmlExists(ctx, amlID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", amlID)
	}

	aml := new(Aml)
	aml.Value = newValue

	bytes, _ := json.Marshal(aml)

	return ctx.GetStub().PutState(amlID, bytes)
}

// DeleteAml deletes an instance of Aml from the world state
func (c *AmlContract) DeleteAml(ctx contractapi.TransactionContextInterface, amlID string) error {
	exists, err := c.AmlExists(ctx, amlID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", amlID)
	}

	return ctx.GetStub().DelState(amlID)
}
