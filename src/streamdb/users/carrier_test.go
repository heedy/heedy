package users

/** Package users provides an API for managing user information.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved
**/

import (
	"testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

	)

func TestCreatePhoneCarrier(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		err := testdb.CreatePhoneCarrier("Test", "example.com")
		require.Nil(t, err, "Cannot create phone carrier %v", err)

		err = testdb.CreatePhoneCarrier("Test", "example2.com")
		assert.NotNil(t, err, "Created carrier with duplicate name")

		err = testdb.CreatePhoneCarrier("Test2", "example.com")
		assert.NotNil(t, err, "Created carrier with duplicate domain")
	}

}

func TestReadAllPhoneCarriers(t *testing.T) {

	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}
		_ = testdb.CreatePhoneCarrier("TestReadAllPhoneCarrier1", "TestReadAllPhoneCarrier1.com")
		_ = testdb.CreatePhoneCarrier("TestReadAllPhoneCarrier2", "TestReadAllPhoneCarrier2.com")

		carriers, err := testdb.ReadAllPhoneCarriers()
		require.Nil(t, err, "Cannot read phone carriers %v", err)

		if len(carriers) < 2 {
			t.Errorf("Did not read all carriers")
		}

		firstfound := false
		secondfound := false
		for _, carrier := range carriers {
			if carrier.Name == "TestReadAllPhoneCarrier1" && carrier.EmailDomain == "TestReadAllPhoneCarrier1.com" {
				firstfound = true
			}

			if carrier.Name == "TestReadAllPhoneCarrier2" && carrier.EmailDomain == "TestReadAllPhoneCarrier2.com" {
				secondfound = true
			}
		}

		if !firstfound {
			t.Errorf("Lost the first carrier")
		}

		if !secondfound {
			t.Errorf("Lost the second carrier")
		}
	}
}

func TestReadPhoneCarrierById(t *testing.T) {
	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}
		err := testdb.CreatePhoneCarrier("TestReadPhoneCarrierById", "TestReadPhoneCarrierById.com")
		assert.Nil(t, err, "Cannot create phone carrier to test")

		pc, err  := testdb.ReadPhoneCarrierByName("TestReadPhoneCarrierById")
		assert.Nil(t, err, "Cannot read carrier by name %v", err)

		id := pc.Id
		carrier, err := testdb.ReadPhoneCarrierById(pc.Id)
		assert.Nil(t, err, "Cannot read phone carrier back with returned id %v", id)
		assert.Equal(t, carrier.Id, id)
		assert.Equal(t, "TestReadPhoneCarrierById", carrier.Name)
		assert.Equal(t, "TestReadPhoneCarrierById.com", carrier.EmailDomain)
	}
}

func TestUpdatePhoneCarrier(t *testing.T) {
	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}
		teststring := "Hello, World!"

		err := testdb.CreatePhoneCarrier("TestUpdatePhoneCarrier", "TestUpdatePhoneCarrier.com")
		assert.Nil(t, err, "Cannot create phone carrier to test")

		pc, err  := testdb.ReadPhoneCarrierByName("TestUpdatePhoneCarrier")
		require.Nil(t, err, "cannot read phone carrier by name")

		id := pc.Id
		carrier, err := testdb.ReadPhoneCarrierById(id)
		require.Nil(t, err, "cannot read phone carrier back with id %d", id)

		carrier.Name = teststring
		err = testdb.UpdatePhoneCarrier(carrier)
		assert.Nil(t, err)

		carrier_back, err := testdb.ReadPhoneCarrierById(id)
		require.Nil(t, err)

		// check if update worked
		assert.Equal(t, teststring, carrier_back.Name)

		err = testdb.UpdatePhoneCarrier(nil)
		require.NotNil(t, err, "updated a nil")
	}
}

func TestDeletePhoneCarrier(t *testing.T) {
	for i, testdb := range(testdatabases) {
		if testdb == nil {
			assert.NotNil(t, testdb, "Could not test database type %v", testdatabasesNames[i])
			continue
		}

		err := testdb.CreatePhoneCarrier("TestDeletePhoneCarrier", "TestDeletePhoneCarrier.com")
		assert.Nil(t, err, "Cannot create phone carrier to test")

		pc, err  := testdb.ReadPhoneCarrierByName("TestDeletePhoneCarrier")
		require.Nil(t, err, "cannot read phone carrier by name")

		id := pc.Id
		err = testdb.DeletePhoneCarrier(id)
		require.Nil(t, err, "Error when attempted delete %v", err)

		_, err = testdb.ReadPhoneCarrierById(id)
		require.NotNil(t, err, "The carrier with the selected ID should have errored out, but it was not")
	}
}
