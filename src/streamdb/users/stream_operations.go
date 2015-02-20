package users

import("database/sql"
"errors"
)



// CreateStream creates a new stream for a given device with the given name, schema and default values.
func (userdb *UserDatabase) CreateStream(Name, Type string, owner *Device) (int64, error) {
    if owner == nil {
        return -1, ERR_INVALID_PTR
    }

    res, err := userdb.db.Exec(`INSERT INTO Stream
        (	Name,
            Type,
            OwnerId) VALUES (?,?,?);`,
            Name, Type, owner.Id)

            if err != nil {
                return 0, err
            }

            return res.LastInsertId()
        }

        // constructStreamsFromRows converts a rows statement to an array of streams
        func constructStreamsFromRows(rows *sql.Rows) ([]*Stream, error) {
            out := []*Stream{}

                // defensive programming
                if rows == nil {
                    return out, ERR_INVALID_PTR
                }

                for rows.Next() {
                    u := new(Stream)
                    err := rows.Scan(
                        &u.Id,
                        &u.Name,
                        &u.Active,
                        &u.Public,
                        &u.Type,
                        &u.OwnerId,
                        &u.Ephemeral,
                        &u.Output)

                        out = append(out, u)

                        if(err != nil) {
                            return out, err
                        }
                    }

                    return out, nil
                }


                // ReadStreamById fetches the stream with the given id and returns it, or nil if
                // no such stream exists.
                func (userdb *UserDatabase) ReadStreamById(id int64) (*Stream, error) {
                    rows, err := userdb.db.Query("SELECT * FROM Stream WHERE Id = ? LIMIT 1", id)

                    if err != nil {
                        return nil, err
                    }

                    defer rows.Close()

                    streams, err := constructStreamsFromRows(rows)

                    if(len(streams) != 1) {
                        return nil, errors.New("Wrong number of streams returned")
                    }

                    return streams[0], nil
                }

                func (userdb *UserDatabase) ReadStreamsByDevice(device *Device) ([]*Stream, error) {
                    rows, err := userdb.db.Query("SELECT * FROM Stream WHERE OwnerId = ? LIMIT 1", device.Id)

                    if err != nil {
                        return nil, err
                    }

                    return constructStreamsFromRows(rows)
                }

                // UpdateStream updates the stream with the given ID with the provided data
                // replacing all prior contents.
                func (userdb *UserDatabase) UpdateStream(stream *Stream) (error) {
                    if stream == nil {
                        return ERR_INVALID_PTR
                    }


                    _, err := userdb.db.Exec(`UPDATE Stream SET
                        Name = ?,
                        Active = ?,
                        Public = ?,
                        Type = ?,
                        OwnerId = ?,
                        Ephemeral = ?,
                        Output = ?
                        WHERE Id = ?;`,
                        stream.Name,
                        stream.Active,
                        stream.Public,
                        stream.Type,
                        stream.OwnerId,
                        stream.Ephemeral,
                        stream.Output,
                        stream.Id)

                        return err
                    }


                    // DeleteStream removes a stream from the database
                    func (userdb *UserDatabase) DeleteStream(Id int64) (error) {
                        _, err := userdb.db.Exec(`DELETE FROM Stream WHERE Id = ?;`, Id );
                        return err
                    }
