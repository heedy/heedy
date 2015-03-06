package users

import "testing"
import "errors"

func TestCreatePhoneCarrier(t *testing.T) {
    _, err := testdb.CreatePhoneCarrier("Test", "example.com")
    if(err != nil) {
        t.Errorf("Cannot create phone carrier %v", err)
        return
    }

    _, err = testdb.CreatePhoneCarrier("Test", "example2.com")
    if(err == nil) {
        t.Errorf("Created carrier with duplicate name")
    }

    _, err = testdb.CreatePhoneCarrier("Test2", "example.com")
    if(err == nil) {
        t.Errorf("Created carrier with duplicate domain")
    }
}

func TestReadAllPhoneCarriers(t *testing.T) {

    _, _ = testdb.CreatePhoneCarrier("TestReadAllPhoneCarrier1", "TestReadAllPhoneCarrier1.com")
    _, _ = testdb.CreatePhoneCarrier("TestReadAllPhoneCarrier2", "TestReadAllPhoneCarrier2.com")

    carriers, err := testdb.ReadAllPhoneCarriers()

    if err != nil {
        t.Errorf("Cannot read phone carriers %v", err)
        return
    }

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

    if ! firstfound {
        t.Errorf("Lost the first carrier")
    }

    if ! secondfound {
        t.Errorf("Lost the second carrier")
    }
}


func TestReadPhoneCarrierById(t *testing.T) {

    id, err := testdb.CreatePhoneCarrier("TestReadPhoneCarrierById", "TestReadPhoneCarrierById.com")

    if nil != err {
        t.Errorf("Cannot create phone carrier to test")
    }

    carrier, err := testdb.ReadPhoneCarrierById(id)

    if err != nil {
        t.Errorf("Cannot read phone carrier back with returned id %v", id)
        return
    }

    if carrier.Id != id {
        t.Errorf("Got mismatching id from carrier, got %v expected %v", carrier.Id, id)
    }

    if carrier.Name != "TestReadPhoneCarrierById" {
        t.Errorf("Got mismatching name from carrier, got %v expected TestReadPhoneCarrierById", carrier.Name)
    }

    if carrier.EmailDomain != "TestReadPhoneCarrierById.com" {
        t.Errorf("Got mismatching name from carrier, got %v expected TestReadPhoneCarrierById.com", carrier.Name)
    }
}

func TestUpdatePhoneCarrier(t *testing.T) {
    teststring := "Hello, World!"

    id, err := testdb.CreatePhoneCarrier("TestUpdatePhoneCarrier", "TestUpdatePhoneCarrier.com")

    if nil != err {
        t.Errorf("Cannot create phone carrier to test")
    }

    carrier, err := testdb.ReadPhoneCarrierById(id)

    if err != nil {
        t.Errorf("Cannot read phone carrier back with returned id %v", id)
        return
    }

    carrier.Name = teststring

    err = testdb.UpdatePhoneCarrier(carrier)

    if err != nil {
        t.Errorf("Cannot update carrier %v", err)
    }

    carrier_back, err := testdb.ReadPhoneCarrierById(id)

    if err != nil {
        t.Errorf("Cannot read phone carrier back with returned id %v", id)
        return
    }

    if carrier_back.Name != teststring {
        t.Errorf("Update did not work, got back %v expected %v", carrier_back.Name, teststring)
    }
}



func TestDeletePhoneCarrier(t *testing.T) {
    id, err := testdb.CreatePhoneCarrier("TestDeletePhoneCarrier", "TestDeletePhoneCarrier.com")

    if nil != err {
        t.Errorf("Cannot create phone carrier to test delete")
        return
    }

    err = testdb.DeletePhoneCarrier(id)

    if nil != err {
        t.Errorf("Error when attempted delete %v", err)
        return
    }

    carrier, err := testdb.ReadPhoneCarrierById(id)

    if err == nil {
        t.Errorf("The carrier with the selected ID should have errored out, but it was not")
        return
    }

    if carrier != nil {
        t.Errorf("Expected nil, but we got back a carrier meaning the delete failed %v", carrier)
    }
}


func TestCarrierSanityChecks(t *testing.T) {

    testerr := errors.New("foobarbaz")

    _, err := constructPhoneCarrierFromRow(nil, testerr)
    if err == nil && err != ERR_INVALID_PTR {
        t.Errorf("Passed in error and got nothing back.")
    }

    if err == ERR_INVALID_PTR {
        t.Errorf("Got a further down error than expecting, should have got the same one back.")
    }

    _, err = constructPhoneCarriersFromRows(nil)

    if err != ERR_INVALID_PTR {
        t.Errorf("Didn't stop a nil pointer")
    }

    err = testdb.UpdatePhoneCarrier(nil)
    if err != ERR_INVALID_PTR {
        t.Errorf("Didn't stop a nil pointer for update")
    }


}
