package users

import("database/sql"
"errors")


// CreatePhoneCarrier creates a phone carrier from the carrier name and
// the SMS email domain they provide, for Example "Tmobile US", "tmomail.net"
func (userdb *UserDatabase) CreatePhoneCarrier(Name, EmailDomain string) (int64, error) {

    res, err := userdb.db.Exec(`INSERT INTO PhoneCarrier (Name, EmailDomain)
    VALUES (?,?);`,
    Name,
    EmailDomain)

    if err != nil {
        return 0, err
    }

    return res.LastInsertId()
}


// constructPhoneCarrierFromRow creates a single PhoneCarrier instance from
// the given rows.
func constructPhoneCarrierFromRow(rows *sql.Rows, err error) (*PhoneCarrier, error) {
    if err != nil {
        return nil, err
    }

    defer rows.Close()

    result, err := constructPhoneCarriersFromRows(rows)

    if err == nil && len(result) > 0 {
        return result[0], err
    }

    return nil, errors.New("No carrier supplied")
}

// constructPhoneCarriersFromRows constructs a series of phone carriers
func constructPhoneCarriersFromRows(rows *sql.Rows) ([]*PhoneCarrier, error) {
    out := []*PhoneCarrier{}

    if rows == nil {
        return out, ERR_INVALID_PTR
    }


    for rows.Next() {
        u := new(PhoneCarrier)
        err := rows.Scan(&u.Id, &u.Name, &u.EmailDomain)

        if err != nil {
            return out, err
        }

        out = append(out, u)
    }

    return out, nil
}

// ReadPhoneCarrierById selects a phone carrier from the database given its ID
func (userdb *UserDatabase) ReadPhoneCarrierById(Id int64) (*PhoneCarrier, error) {
    rows, err := userdb.db.Query("SELECT * FROM PhoneCarrier WHERE Id = ? LIMIT 1", Id)
    return constructPhoneCarrierFromRow(rows, err)
}

func (userdb *UserDatabase) ReadAllPhoneCarriers() ([]*PhoneCarrier, error) {
    rows, err := userdb.db.Query("SELECT * FROM PhoneCarrier")

    if err != nil {
        return nil, err
    }

    defer rows.Close()
    return constructPhoneCarriersFromRows(rows)
}

// UpdatePhoneCarrier updates the database's phone carrier data with that of the
// struct provided.
func (userdb *UserDatabase) UpdatePhoneCarrier(carrier *PhoneCarrier) (error) {
    if carrier == nil {
        return ERR_INVALID_PTR
    }


    _, err := userdb.db.Exec(`UPDATE PhoneCarrier SET
        Name=?, EmailDomain=? WHERE Id = ?;`,
        carrier.Name,
        carrier.EmailDomain,
        carrier.Id)

    return err
}

// DeletePhoneCarrier removes a phone carrier from the database, this will set
// all users carrier with this phone carrier as a foreign key to NULL
func (userdb *UserDatabase) DeletePhoneCarrier(carrierId int64) (error) {
    _, err := userdb.db.Exec(`DELETE FROM PhoneCarrier WHERE Id = ?;`, carrierId )
    return err
}
