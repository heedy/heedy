package users

import "testing"

func TestCreatePhoneCarrier(t *testing.T) {
    _, err := CreatePhoneCarrier("Test", "example.com")
    if(err != nil) {
        t.Errorf("Cannot create phone carrier %v", err)
        return
    }

    _, err = CreatePhoneCarrier("Test", "example2.com")
    if(err == nil) {
        t.Errorf("Created carrier with duplicate name")
    }

    _, err = CreatePhoneCarrier("Test2", "example.com")
    if(err == nil) {
        t.Errorf("Created carrier with duplicate domain")
    }
}

func TestReadAllPhoneCarriers(t *testing.T) {

    _, _ = CreatePhoneCarrier("TestReadAllPhoneCarrier1", "TestReadAllPhoneCarrier1.com")
    _, _ = CreatePhoneCarrier("TestReadAllPhoneCarrier2", "TestReadAllPhoneCarrier2.com")

    carriers, err := ReadAllPhoneCarriers()

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

    id, err := CreatePhoneCarrier("TestReadPhoneCarrierById", "TestReadPhoneCarrierById.com")

    if nil != err {
        t.Errorf("Cannot create phone carrier to test")
    }

    carrier, err := ReadPhoneCarrierById(id)

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

    id, err := CreatePhoneCarrier("TestUpdatePhoneCarrier", "TestUpdatePhoneCarrier.com")

    if nil != err {
        t.Errorf("Cannot create phone carrier to test")
    }

    carrier, err := ReadPhoneCarrierById(id)

    if err != nil {
        t.Errorf("Cannot read phone carrier back with returned id %v", id)
        return
    }

    carrier.Name = teststring

    err = UpdatePhoneCarrier(carrier)

    if err != nil {
        t.Errorf("Cannot update carrier %v", err)
    }

    carrier_back, err := ReadPhoneCarrierById(id)

    if err != nil {
        t.Errorf("Cannot read phone carrier back with returned id %v", id)
        return
    }

    if carrier_back.Name != teststring {
        t.Errorf("Update did not work, got back %v expected %v", carrier_back.Name, teststring)
    }
}



func TestDeletePhoneCarrier(t *testing.T) {
    id, err := CreatePhoneCarrier("TestDeletePhoneCarrier", "TestDeletePhoneCarrier.com")

    if nil != err {
        t.Errorf("Cannot create phone carrier to test delete")
        return
    }

    err = DeletePhoneCarrier(id)

    if nil != err {
        t.Errorf("Error when attempted delete %v", err)
        return
    }

    carrier, err := ReadPhoneCarrierById(id)

    if err == nil {
        t.Errorf("The carrier with the selected ID should have errored out, but it was not")
        return
    }

    if carrier != nil {
        t.Errorf("Expected nil, but we got back a carrier meaning the delete failed %v", carrier)
    }
}
