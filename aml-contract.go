/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AmlContract contract for managing CRUD for Aml
type AmlContract struct {
	contractapi.Contract
}

func (c *AmlContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	amlData := []Aml{
		Aml{Last_name: "Lee", First_name: "Tom", DOB: "1980/01/02", Country: "TWN", ID_number: "A123456789", Data_owner: "org0MSP", Risk_level: "low"},
		Aml{Last_name: "Tseng", First_name: "Ling-Pei", DOB: "1982/02/20", Country: "TWN", ID_number: "D111111111", Data_owner: "org0MSP", Risk_level: "high"},
		Aml{Last_name: "Chan", First_name: "Yip", DOB: "1970/02/15", Country: "HKG", ID_number: "ABZG156465", Data_owner: "org0MSP", Risk_level: "medium"},

		Aml{Last_name: "Lee", First_name: "Tom", DOB: "1980/01/02", Country: "TWN", ID_number: "A123456789", Data_owner: "org1MSP", Risk_level: "low"},
		Aml{Last_name: "Li", First_name: "Kuei-Jung", DOB: "1973/10/04", Country: "NLD", ID_number: "CALZ12557", Data_owner: "org1MSP", Risk_level: "low"},
		Aml{Last_name: "Shen", First_name: "Lung-Tsu", DOB: "1979/05/14", Country: "TWN", ID_number: "F123456789", Data_owner: "org1MSP", Risk_level: "low"},

		Aml{Last_name: "Lee", First_name: "Tom", DOB: "1980/01/02", Country: "TWN", ID_number: "A123456789", Data_owner: "org2MSP", Risk_level: "low"},
		Aml{Last_name: "TSUNG", First_name: "CHUN-CHEN", DOB: "1982/06/10", Country: "TWN", ID_number: "B123456789", Data_owner: "org2MSP", Risk_level: "medium"},
		Aml{Last_name: "Chan", First_name: "Chi-Jong", DOB: "1975/04/03", Country: "TWN", ID_number: "C123456789", Data_owner: "org2MSP", Risk_level: "low"},
	}
	for _, aml := range amlData {
		AmlAsbytes, _ := json.Marshal(aml)
		key := aml.Country + "_" + aml.ID_number + "_" + aml.Data_owner
		err := ctx.GetStub().PutState(key, AmlAsbytes)

		if err != nil {
			return fmt.Errorf("failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

// AmlExists returns true when asset with given ID exists in world state
func (c *AmlContract) AmlExists(ctx contractapi.TransactionContextInterface, key string) (bool, error) {

	data, err := ctx.GetStub().GetState(key)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateAml creates a new instance of Aml
func (c *AmlContract) CreateAmlData(ctx contractapi.TransactionContextInterface, last_name string, first_name string, dob string, country string, id_number string, risk_level string) error {

	// Get client org id and verify it matches peer org id.
	// In this scenario, client is only authorized to read/write private data from its own peer.
	clientOrgID, err := getClientOrgID(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to get verified OrgID: %v", err)
	}

	key := country + "_" + id_number + "_" + clientOrgID
	exists, err := c.AmlExists(ctx, key)
	if err != nil {
		return fmt.Errorf("could not interact with aml world state. %s", err)
	} else if exists {
		return fmt.Errorf("the aml data already exists country:%s, id_number:%s, data_owner:%s", country, id_number, clientOrgID)
	}

	aml := new(Aml)
	aml.Last_name = last_name
	aml.First_name = first_name
	aml.DOB = dob
	aml.Country = country
	aml.ID_number = id_number
	aml.Data_owner = clientOrgID
	aml.Risk_level = risk_level

	bytes, _ := json.Marshal(aml)

	err = ctx.GetStub().PutState(key, bytes)
	if err != nil {
		return fmt.Errorf("could not interact with aml world state. %s", err)
	}

	// Set the endorsement policy such that an owner org peer is required to endorse future updates
	err = setAssetStateBasedEndorsement(ctx, key, clientOrgID)
	if err != nil {
		return fmt.Errorf("failed setting state based endorsement for owner: %v", err)
	}

	return nil
}

// getClientOrgID gets the client org ID.
// The client org ID can optionally be verified against the peer org ID, to ensure that a client
// from another org doesn't attempt to read or write private data from this peer.
// The only exception in this scenario is for TransferAsset, since the current owner
// needs to get an endorsement from the buyer's peer.
func getClientOrgID(ctx contractapi.TransactionContextInterface, verifyOrg bool) (string, error) {
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}

	if verifyOrg {
		err = verifyClientOrgMatchesPeerOrg(clientOrgID)
		if err != nil {
			return "", err
		}
	}

	return clientOrgID, nil
}

// verifyClientOrgMatchesPeerOrg checks the client org id matches the peer org id.
func verifyClientOrgMatchesPeerOrg(clientOrgID string) error {
	peerOrgID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting peer's orgID: %v", err)
	}

	if clientOrgID != peerOrgID {
		return fmt.Errorf("client from org %s is not authorized to read or write private data from an org %s peer",
			clientOrgID,
			peerOrgID,
		)
	}

	return nil
}

// setAssetStateBasedEndorsement adds an endorsement policy to a asset so that only a peer from an owning org
// can update or transfer the asset.
func setAssetStateBasedEndorsement(ctx contractapi.TransactionContextInterface, key string, orgToEndorse string) error {
	endorsementPolicy, err := statebased.NewStateEP(nil)
	if err != nil {
		return err
	}
	err = endorsementPolicy.AddOrgs(statebased.RoleTypePeer, orgToEndorse)
	if err != nil {
		return fmt.Errorf("failed to add org to endorsement policy: %v", err)
	}
	policy, err := endorsementPolicy.Policy()
	if err != nil {
		return fmt.Errorf("failed to create endorsement policy bytes from org: %v", err)
	}
	err = ctx.GetStub().SetStateValidationParameter(key, policy)
	if err != nil {
		return fmt.Errorf("failed to set validation parameter on asset: %v", err)
	}

	return nil
}

func (t *AmlContract) Query(ctx contractapi.TransactionContextInterface, queryString string) ([]*Aml, error) {
	return getQueryResultForQueryString(ctx, queryString)
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Aml, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// constructQueryResponseFromIterator constructs a slice of assets from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Aml, error) {
	var assets []*Aml
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset Aml
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// UpdateAml retrieves an instance of Aml from the world state and updates its value
func (c *AmlContract) UpdateAmlData(ctx contractapi.TransactionContextInterface, last_name string, first_name string, dob string, country string, id_number string, risk_level string) error {

	// No need to check client org id matches peer org id, rely on the asset ownership check instead.
	clientOrgID, err := getClientOrgID(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get verified OrgID: %v", err)
	}

	key := country + "_" + id_number + "_" + clientOrgID
	exists, err := c.AmlExists(ctx, key)
	if err != nil {
		return fmt.Errorf("could not interact with aml world state. %s", err)
	} else if !exists {
		return fmt.Errorf("the aml data does not exists: country:%s, id_number:%s, data_owner:%s", country, id_number, clientOrgID)
	}

	aml := new(Aml)
	aml.Last_name = last_name
	aml.First_name = first_name
	aml.DOB = dob
	aml.Country = country
	aml.ID_number = id_number
	aml.Data_owner = clientOrgID
	aml.Risk_level = risk_level

	bytes, _ := json.Marshal(aml)

	return ctx.GetStub().PutState(key, bytes)
}

// DeleteAml deletes an instance of Aml from the world state
func (c *AmlContract) DeleteAmlData(ctx contractapi.TransactionContextInterface, country string, id_number string) error {
	// No need to check client org id matches peer org id, rely on the asset ownership check instead.
	clientOrgID, err := getClientOrgID(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get verified OrgID: %v", err)
	}

	key := country + "_" + id_number + "_" + clientOrgID
	exists, err := c.AmlExists(ctx, key)
	if err != nil {
		return fmt.Errorf("could not interact with aml world state. %s", err)
	} else if !exists {
		return fmt.Errorf("the aml data does not exist, country:%s, id_number:%s, data_owner:%s", country, id_number, clientOrgID)
	}

	return ctx.GetStub().DelState(key)
}

// GetAssetHistory returns the chain of custody for an asset since issuance.
func (c *AmlContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, country string, id_number string, data_owner string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: Country: %s, ID_number: %s, Data_owner: %s", country, id_number, data_owner)
	key := country + "_" + id_number + "_" + data_owner
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Aml
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Aml{}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}
