/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package datastream

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
)

/*
The datastream table:

CREATE TABLE IF NOT EXISTS datastream (
    StreamID BIGINT NOT NULL,
	Substream VARCHAR,
    EndTime DOUBLE PRECISION,
    EndIndex BIGINT,
	Version INTEGER,
    Data BYTEA,
    UNIQUE (StreamID, Substream, EndIndex),
    PRIMARY KEY (StreamID, Substream, EndIndex)
    );
*/

var (
	//ErrorDatabaseCorrupted is returned when there is data loss or inconsistency in the database
	ErrorDatabaseCorrupted = errors.New("Database is corrupted!")
	//ErrWTF is returned when an internal assertion fails - it shoudl not happen. Ever.
	ErrWTF = errors.New("Something is seriously wrong. A internal assertion failed.")
)

//The SqlStore stores and queries arrays of Datapoints in an SQL database. The table 'datastream' is assumed
//to already exist and the correct indices are assumed to already exist.
type SqlStore struct {
	inserter     *sqlx.Stmt
	timequery    *sqlx.Stmt
	indexquery   *sqlx.Stmt
	endindex     *sqlx.Stmt
	delsubstream *sqlx.Stmt
	delstream    *sqlx.Stmt
	clearall     *sqlx.Stmt

	db *sqlx.DB

	insertversion int
}

//This function is to allow daisy-chaining errors from statement creation
func prepStatement(db *sqlx.DB, statement string, err error) (*sqlx.Stmt, error) {
	if err != nil {
		return nil, err
	}
	return db.Preparex(db.Rebind(statement))
}

//OpenSqlStore initializes the database statements
func OpenSqlStore(db *sqlx.DB) (*SqlStore, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}

	inserter, err := prepStatement(db, "INSERT INTO datastream VALUES (?,?,?,?,?,?);", nil)
	timequery, err := prepStatement(db, "SELECT version,endindex,data FROM datastream WHERE streamid=? AND substream=? AND endtime > ? ORDER BY endtime ASC;", err)
	indexquery, err := prepStatement(db, "SELECT version,endindex,data FROM datastream WHERE streamid=? AND substream=? AND endindex > ? ORDER BY endindex ASC;", err)
	endindex, err := prepStatement(db, "SELECT COALESCE(MAX(endindex),0) FROM datastream WHERE streamid=? AND substream=?;", err)
	delsubstream, err := prepStatement(db, "DELETE FROM datastream WHERE streamid=? AND substream=?;", err)
	delstream, err := prepStatement(db, "DELETE FROM datastream WHERE streamid=?;", err)
	clearall, err := prepStatement(db, "DELETE FROM datastream;", err)

	ss := &SqlStore{inserter, timequery, indexquery, endindex, delsubstream, delstream, clearall, db, 2}

	if err != nil {
		ss.Close()
		return nil, err
	}

	return ss, nil
}

//Close all resources associated with the SqlStore.
func (s *SqlStore) Close() {
	//The if statements allow to close a partially initialized store
	if s.inserter != nil {
		s.inserter.Close()
	}
	if s.timequery != nil {
		s.timequery.Close()
	}
	if s.indexquery != nil {
		s.indexquery.Close()
	}
	if s.endindex != nil {
		s.endindex.Close()
	}
	if s.delstream != nil {
		s.delstream.Close()
	}
	if s.delsubstream != nil {
		s.delsubstream.Close()
	}
}

//Clear the entire table of all data
func (s *SqlStore) Clear() error {
	_, err := s.clearall.Exec()
	return err
}

//GetEndIndex returns the first index point outside of the most recent datapointarray stored within the database.
//In effect, if the datapoints in a key were all in one huge array, returns array.length
//(not including the datapoints which are not yet committed to the SqlStore)
func (s *SqlStore) GetEndIndex(streamID int64, substream string) (ei int64, err error) {
	rows, err := s.endindex.Query(streamID, substream)
	if err != nil {
		return 0, err
	}
	if !rows.Next() {
		return 0, ErrWTF //This should never happen
	}
	err = rows.Scan(&ei)
	rows.Close()
	return ei, err
}

//Insert the given DatapointArray into the sql database given the startindex of the array for the key.
func (s *SqlStore) Insert(streamID int64, substream string, startindex int64, da DatapointArray) error {
	return s.stmtInsert(s.inserter, streamID, substream, startindex, da)
}

func (s *SqlStore) stmtInsert(stmt *sqlx.Stmt, streamID int64, substream string, startindex int64, da DatapointArray) error {
	dbytes, err := da.Encode(s.insertversion)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(streamID, substream, da[len(da)-1].Timestamp, startindex+int64(len(da)),
		s.insertversion, dbytes)
	return err
}

//WriteBatches writes the given batch array
func (s *SqlStore) WriteBatches(b []Batch) error {
	t, err := s.db.Beginx()
	if err != nil {
		return err
	}

	for i := 0; i < len(b); i++ {
		log.Debugf("Writing batch %s/%s i=%d #=%d", b[i].Stream, b[i].Substream, b[i].StartIndex, len(b[i].Data))
		streamID, err := b[i].GetStreamID()
		if err != nil {
			t.Rollback()
			return err
		}

		//Now the transaction-specific insert statement
		err = s.stmtInsert(t.Stmtx(s.inserter), streamID, b[i].Substream, b[i].StartIndex, b[i].Data)
		if err != nil {
			t.Rollback()
			return err
		}
	}
	err = t.Commit()
	if err == nil && len(b) > 1 {
		log.Debugf("...successfully wrote %d batches", len(b))
	}
	return err
}

//Append the given DatapointArray to the data stream for key
func (s *SqlStore) Append(streamID int64, substream string, dp DatapointArray) error {
	i, err := s.GetEndIndex(streamID, substream)
	if err != nil {
		return err
	}
	return s.Insert(streamID, substream, i, dp)
}

//DeleteStream deletes all data associated with the given stream in the database
func (s *SqlStore) DeleteStream(streamID int64) error {
	_, err := s.delstream.Exec(streamID)
	return err
}

//DeleteSubstream deletes all data associated with the given substream in the database
func (s *SqlStore) DeleteSubstream(streamID int64, substream string) error {
	_, err := s.delsubstream.Exec(streamID, substream)
	return err
}

//GetByTime returns a ExtendedDataRange of datapoints starting at the starttime
func (s *SqlStore) GetByTime(streamID int64, substream string, starttime float64) (dr ExtendedDataRange, startindex int64, err error) {
	rows, err := s.timequery.Query(streamID, substream, starttime)
	if err != nil {
		return nil, 0, err
	}

	if !rows.Next() { //Check if there is any data to read
		startindex, err = s.GetEndIndex(streamID, substream)
		if rows.Err() != nil {
			err = rows.Err()
		}
		rows.Close()
		return EmptyRange{}, startindex, err
	}

	//There is some data!
	var version int
	var endindex int64
	var data []byte
	if err = rows.Scan(&version, &endindex, &data); err != nil {
		return EmptyRange{}, endindex, err
	}

	da, err := DecodeDatapointArray(data, version)
	if err != nil {
		rows.Close()
		return EmptyRange{}, endindex, err
	}
	tmp := da.TStart(starttime)
	da = &tmp
	if da == nil || int64(da.Length()) > endindex {
		rows.Close()
		return EmptyRange{}, endindex, ErrorDatabaseCorrupted
	}
	curindex := endindex - int64(da.Length())
	return &SqlRange{rows, da, curindex}, curindex, nil
}

//GetByIndex returns a ExtendedDataRange of datapoints starting at the nearest dataindex to the given startindex
func (s *SqlStore) GetByIndex(streamID int64, substream string, startindex int64) (dr ExtendedDataRange, dataindex int64, err error) {
	rows, err := s.indexquery.Query(streamID, substream, startindex)
	if err != nil {
		return nil, 0, err
	}

	if !rows.Next() { //Check if there is any data to read
		startindex, err = s.GetEndIndex(streamID, substream)
		if rows.Err() != nil {
			err = rows.Err()
		}
		rows.Close()
		return EmptyRange{}, startindex, err
	}

	//There is some data!
	var version int
	var endindex int64
	var data []byte
	if err = rows.Scan(&version, &endindex, &data); err != nil {
		return EmptyRange{}, endindex, err
	}

	da, err := DecodeDatapointArray(data, version)
	if err != nil {
		rows.Close()
		return EmptyRange{}, endindex, err
	}

	if da == nil || int64(da.Length()) > endindex {
		rows.Close()
		return EmptyRange{}, endindex, ErrorDatabaseCorrupted
	}

	//Lastly, we start the DatapointArray from the correct index
	//This subtraction is guaranteed to work, since query requires $gt
	fromend := endindex - startindex
	if fromend < int64(da.Length()) {
		//The index we want is within the datarange
		da = da.IRange(da.Length()-int(fromend), da.Length())
	}
	curindex := endindex - int64(da.Length())
	return &SqlRange{rows, da, curindex}, curindex, nil
}
