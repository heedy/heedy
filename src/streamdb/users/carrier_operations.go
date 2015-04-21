package users

/** Package users provides an API for managing user information.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved
**/


// A phone carrier is a mobile service provider that has email forwarding of
// its
type PhoneCarrier struct {
	Id          int64  `modifiable:"nobody"`
	Name        string `modifiable:"root"`
	EmailDomain string `modifiable:"root"`
}


// CreatePhoneCarrier creates a phone carrier from the carrier name and
// the SMS email domain they provide, for Example "Tmobile US", "tmomail.net"
func (userdb *UserDatabase) CreatePhoneCarrier(Name, EmailDomain string) (error) {
	_, err := userdb.Exec(`INSERT INTO PhoneCarriers (Name, EmailDomain) VALUES (?,?);`,
		Name,
		EmailDomain)

	if err != nil {
		return err
	}

	return err
}

// ReadPhoneCarrierById selects a phone carrier from the database given its ID
func (userdb *UserDatabase) ReadPhoneCarrierById(Id int64) (*PhoneCarrier, error) {
	var pc PhoneCarrier

	err := userdb.Get(&pc, "SELECT * FROM PhoneCarriers WHERE Id = ? LIMIT 1", Id)

	return &pc, err
}

// ReadPhoneCarrierById selects a phone carrier from the database given its ID
func (userdb *UserDatabase) ReadPhoneCarrierByName(name string) (*PhoneCarrier, error) {
	var pc PhoneCarrier

	err := userdb.Get(&pc, "SELECT * FROM PhoneCarriers WHERE Name = ? LIMIT 1", name)

	return &pc, err
}

func (userdb *UserDatabase) ReadAllPhoneCarriers() ([]PhoneCarrier, error) {
	var carriers []PhoneCarrier

	err := userdb.Select(&carriers, "SELECT * FROM PhoneCarriers")

	return carriers, err
}

// UpdatePhoneCarrier updates the database's phone carrier data with that of the
// struct provided.
func (userdb *UserDatabase) UpdatePhoneCarrier(carrier *PhoneCarrier) error {
	if carrier == nil {
		return ERR_INVALID_PTR
	}

	_, err := userdb.Exec(`UPDATE PhoneCarriers SET Name=?, EmailDomain=? WHERE Id = ?;`,
		carrier.Name,
		carrier.EmailDomain,
		carrier.Id)

	return err
}

// DeletePhoneCarrier removes a phone carrier from the database, this will set
// all users carrier with this phone carrier as a foreign key to NULL
func (userdb *UserDatabase) DeletePhoneCarrier(carrierId int64) error {
	_, err := userdb.Exec(`DELETE FROM PhoneCarriers WHERE Id = ?;`, carrierId)

	return err
}
