/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const getStateError = "world state get error"

type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
}

func (ms *MockStub) GetState(key string) ([]byte, error) {
	args := ms.Called(key)

	return args.Get(0).([]byte), args.Error(1)
}

func (ms *MockStub) PutState(key string, value []byte) error {
	args := ms.Called(key, value)

	return args.Error(0)
}

func (ms *MockStub) DelState(key string) error {
	args := ms.Called(key)

	return args.Error(0)
}

type MockContext struct {
	contractapi.TransactionContextInterface
	mock.Mock
}

func (mc *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := mc.Called()

	return args.Get(0).(*MockStub)
}

func configureStub() (*MockContext, *MockStub) {
	var nilBytes []byte

	testAml := new(Aml)
	testAml.Value = "set value"
	amlBytes, _ := json.Marshal(testAml)

	ms := new(MockStub)
	ms.On("GetState", "statebad").Return(nilBytes, errors.New(getStateError))
	ms.On("GetState", "missingkey").Return(nilBytes, nil)
	ms.On("GetState", "existingkey").Return([]byte("some value"), nil)
	ms.On("GetState", "amlkey").Return(amlBytes, nil)
	ms.On("PutState", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
	ms.On("DelState", mock.AnythingOfType("string")).Return(nil)

	mc := new(MockContext)
	mc.On("GetStub").Return(ms)

	return mc, ms
}

func TestAmlExists(t *testing.T) {
	var exists bool
	var err error

	ctx, _ := configureStub()
	c := new(AmlContract)

	exists, err = c.AmlExists(ctx, "statebad")
	assert.EqualError(t, err, getStateError)
	assert.False(t, exists, "should return false on error")

	exists, err = c.AmlExists(ctx, "missingkey")
	assert.Nil(t, err, "should not return error when can read from world state but no value for key")
	assert.False(t, exists, "should return false when no value for key in world state")

	exists, err = c.AmlExists(ctx, "existingkey")
	assert.Nil(t, err, "should not return error when can read from world state and value exists for key")
	assert.True(t, exists, "should return true when value for key in world state")
}

func TestCreateAml(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	c := new(AmlContract)

	err = c.CreateAml(ctx, "statebad", "some value")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	err = c.CreateAml(ctx, "existingkey", "some value")
	assert.EqualError(t, err, "The asset existingkey already exists", "should error when exists returns true")

	err = c.CreateAml(ctx, "missingkey", "some value")
	stub.AssertCalled(t, "PutState", "missingkey", []byte("{\"value\":\"some value\"}"))
}

func TestReadAml(t *testing.T) {
	var aml *Aml
	var err error

	ctx, _ := configureStub()
	c := new(AmlContract)

	aml, err = c.ReadAml(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors when reading")
	assert.Nil(t, aml, "should not return Aml when exists errors when reading")

	aml, err = c.ReadAml(ctx, "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when reading")
	assert.Nil(t, aml, "should not return Aml when key does not exist in world state when reading")

	aml, err = c.ReadAml(ctx, "existingkey")
	assert.EqualError(t, err, "Could not unmarshal world state data to type Aml", "should error when data in key is not Aml")
	assert.Nil(t, aml, "should not return Aml when data in key is not of type Aml")

	aml, err = c.ReadAml(ctx, "amlkey")
	expectedAml := new(Aml)
	expectedAml.Value = "set value"
	assert.Nil(t, err, "should not return error when Aml exists in world state when reading")
	assert.Equal(t, expectedAml, aml, "should return deserialized Aml from world state")
}

func TestUpdateAml(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	c := new(AmlContract)

	err = c.UpdateAml(ctx, "statebad", "new value")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors when updating")

	err = c.UpdateAml(ctx, "missingkey", "new value")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when updating")

	err = c.UpdateAml(ctx, "amlkey", "new value")
	expectedAml := new(Aml)
	expectedAml.Value = "new value"
	expectedAmlBytes, _ := json.Marshal(expectedAml)
	assert.Nil(t, err, "should not return error when Aml exists in world state when updating")
	stub.AssertCalled(t, "PutState", "amlkey", expectedAmlBytes)
}

func TestDeleteAml(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	c := new(AmlContract)

	err = c.DeleteAml(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("Could not read from world state. %s", getStateError), "should error when exists errors")

	err = c.DeleteAml(ctx, "missingkey")
	assert.EqualError(t, err, "The asset missingkey does not exist", "should error when exists returns true when deleting")

	err = c.DeleteAml(ctx, "amlkey")
	assert.Nil(t, err, "should not return error when Aml exists in world state when deleting")
	stub.AssertCalled(t, "DelState", "amlkey")
}
