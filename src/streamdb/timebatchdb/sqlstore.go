package timebatchdb

import (
	"database/sql"
	"errors"
)

var (
	ERROR_DATABASE_DRIVER    = errors.New("Database driver not supported")
	ERROR_DATABASE_CORRUPTED = errors.New("Database is corrupted!")
	ERROR_WTF                = errors.New("Something is seriously wrong. A internal assertion failed.")
)

//The DataRange which handles retrieving data from an Sql database
type SqlRange struct {
	r  *sql.Rows
	da *DatapointArray
}

//Clears all resources used by the SqlRange
func (s *SqlRange) Close() {
	if s.r != nil {
		s.r.Close()
		s.r = nil
	}
}

//A dummy function, it doesn't actually do anything.
//It exists just so that SqlRange fits the DataRange interface
func (s *SqlRange) Init() error {
	return nil
}

//Returns the next datapoint from the SqlRange
func (s *SqlRange) Next() (*Datapoint, error) {
	d, _ := s.da.Next() //Next on DatapointArray never returns error
	if d != nil {
		return d, nil
	}

	//The DatapointArray is now empty - check if the iterator is still functional
	if s.r == nil {
		return nil, nil
	}

	if !s.r.Next() { //Check if there is more data to read
		err := s.r.Err()
		s.Close()
		return nil, err
	}

	//There is more data to read!
	var endindex uint64 //We don't actually care about this in our case - but we need to scan it
	var data []byte
	if err := s.r.Scan(&endindex, &data); err != nil {
		s.Close()
		return nil, err
	}
	s.da = DatapointArrayFromBytes(data)
	//s.da = DatapointArrayFromCompressedBytes(data)

	//Repeat the procedure.
	return s.Next()
}

//The SqlStore stores and queries arrays of Datapoints in an SQL database. The table 'timebatchtable' is assumed
//to already exist and the correct indices are assumed to already exist.
type SqlStore struct {
	inserter   *sql.Stmt
	timequery  *sql.Stmt
	indexquery *sql.Stmt
	endindex   *sql.Stmt
	delkey     *sql.Stmt
}

//Closes all resources associated with the SqlStore.
func (s *SqlStore) Close() {
	s.inserter.Close()
	s.timequery.Close()
	s.indexquery.Close()
	s.endindex.Close()
	s.delkey.Close()
}

//Returns the first index point outside of the most recent datapointarray stored within the database.
//In effect, if the datapoints in a key were all in one huge array, returns array.length
//(not including the datapoints which are not yet committed to the SqlStore)
func (s *SqlStore) GetEndIndex(key string) (ei uint64, err error) {
	rows, err := s.endindex.Query(key)
	if err != nil {
		return 0, err
	}
	if !rows.Next() {
		return 0, ERROR_WTF //This should never happen
	}
	err = rows.Scan(&ei)
	rows.Close()
	return ei, err
}

//Inserts the given DatapointArray into the sql database given the startindex of the array for the key.
func (s *SqlStore) Insert(key string, startindex uint64, da *DatapointArray) error {
	_, err := s.inserter.Exec(key, da.Datapoints[da.Len()-1].Timestamp(),
		startindex+uint64(da.Len()), da.Bytes())
	//startindex+uint64(da.Len()),da.CompressedBytes())
	return err
}

//Appends the given DatapointArray to the data stream for key
func (s *SqlStore) Append(key string, dp *DatapointArray) error {
	i, err := s.GetEndIndex(key)
	if err != nil {
		return err
	}
	return s.Insert(key, i, dp)
}

//Deletes all data associated with the given key in the database
func (s *SqlStore) Delete(key string) error {
	_, err := s.delkey.Exec(key)
	return err
}

//Returns an SqlRange of datapoints starting at the starttime
func (s *SqlStore) GetByTime(key string, starttime int64) (dr DataRange, startindex uint64, err error) {
	rows, err := s.timequery.Query(key, starttime)
	if err != nil {
		return EmptyRange{}, 0, err
	}

	if !rows.Next() { //Check if there is any data to read
		startindex, err = s.GetEndIndex(key)
		if rows.Err() != nil {
			err = rows.Err()
		}
		return EmptyRange{}, startindex, rows.Err()
	}

	//There is some data!
	var endindex uint64
	var data []byte
	if err = rows.Scan(&endindex, &data); err != nil {
		return EmptyRange{}, endindex, err
	}

	da := DatapointArrayFromBytes(data).TStart(starttime)
	//da := DatapointArrayFromCompressedBytes(data).TStart(starttime)
	if da == nil || uint64(da.Len()) > endindex {
		rows.Close()
		return EmptyRange{}, endindex, ERROR_DATABASE_CORRUPTED
	}

	return &SqlRange{rows, da}, endindex - uint64(da.Len()), nil
}

//Returns an SqlRange of datapoints starting at the nearest dataindex to the given startindex
func (s *SqlStore) GetByIndex(key string, startindex uint64) (dr DataRange, dataindex uint64, err error) {
	rows, err := s.indexquery.Query(key, startindex)
	if err != nil {
		return EmptyRange{}, 0, err
	}

	if !rows.Next() { //Check if there is any data to read
		dataindex, err = s.GetEndIndex(key)
		if rows.Err() != nil {
			err = rows.Err()
		}
		return EmptyRange{}, dataindex, rows.Err()
	}

	//There is some data!
	var endindex uint64
	var data []byte
	if err = rows.Scan(&endindex, &data); err != nil {
		return EmptyRange{}, endindex, err
	}
	da := DatapointArrayFromBytes(data)
	//da := DatapointArrayFromCompressedBytes(data)

	if da == nil || uint64(da.Len()) > endindex {
		rows.Close()
		return EmptyRange{}, endindex, ERROR_DATABASE_CORRUPTED
	}

	//Lastly, we start the DatapointArray from the correct index
	//This subtraction is guaranteed to work on uint, since query requires $gt
	fromend := endindex - startindex
	if fromend < uint64(da.Len()) {
		//The index we want is within the datarange
		da = NewDatapointArray(da.Datapoints[da.Len()-int(fromend):])
	}

	return &SqlRange{rows, da}, endindex - uint64(da.Len()), nil
}

//Sets up the inserts (it assumes that the database was already prepared)
func prepareSqlStore(db *sql.DB, insert_s, timequery_s, indexquery_s, endindex_s string, delkey_s string) (*SqlStore, error) {
	inserter, err := db.Prepare(insert_s)
	if err != nil {
		return nil, err
	}

	timequery, err := db.Prepare(timequery_s)
	if err != nil {
		inserter.Close()
		return nil, err
	}

	indexquery, err := db.Prepare(indexquery_s)
	if err != nil {
		inserter.Close()
		timequery.Close()
		return nil, err
	}

	endindex, err := db.Prepare(endindex_s)
	if err != nil {
		inserter.Close()
		timequery.Close()
		indexquery.Close()
		return nil, err
	}

	delkey, err := db.Prepare(delkey_s)
	if err != nil {
		inserter.Close()
		timequery.Close()
		indexquery.Close()
		endindex.Close()
		return nil, err
	}

	return &SqlStore{inserter, timequery, indexquery, endindex, delkey}, nil
}

//Initializes an sqlite database to work with an SqlStore.
func OpenSQLiteStore(db *sql.DB) (*SqlStore, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS timebatchtable
        (
            Key STRING NOT NULL,
            EndTime INTEGER,
            EndIndex INTEGER,
            Data BLOB,
            PRIMARY KEY (Key, EndIndex)
            );`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec("CREATE INDEX IF NOT EXISTS KeyTime ON timebatchtable (Key,EndTime ASC)")
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	//Now that tables are all set up, prepare the queries to run on the database

	return prepareSqlStore(db, "INSERT INTO timebatchtable VALUES (?,?,?,?);",
		"SELECT EndIndex,Data FROM timebatchtable WHERE Key=? AND EndTime > ? ORDER BY EndTime ASC",
		"SELECT EndIndex,Data FROM timebatchtable WHERE Key=? AND EndIndex > ? ORDER BY EndIndex ASC",
		"SELECT ifnull(max(EndIndex),0) FROM timebatchtable WHERE Key=?",
		"DELETE FROM timebatchtable WHERE Key=?")
}

//Initializes an sqlite database to work with an SqlStore.
func OpenPostgresStore(db *sql.DB) (*SqlStore, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS timebatchtable
        (
            Key VARCHAR NOT NULL,
            EndTime BIGINT,
            EndIndex BIGINT,
            Data BYTEA,
            PRIMARY KEY (Key, EndIndex)
            );`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	//As of writing this code, postgres does not have CREATE INDEX IF NOT EXISTS, so we manually check it.
	//Based upon http://stackoverflow.com/questions/24674281/create-unique-index-if-not-exists-in-postgresql
	_, err = tx.Exec(`DO
        $$
        DECLARE
            l_count integer;
        BEGIN
            select count(*) into l_count from pg_indexes where schemaname = 'public'
                        and tablename = 'timebatchtable' and indexname = 'keytime';
            if l_count = 0 then
                CREATE INDEX keytime ON timebatchtable (Key,EndTime ASC);
            end if;
        END;
        $$`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	//Now that tables are all set up, prepare the queries to run on the database
	return prepareSqlStore(db, "INSERT INTO timebatchtable VALUES ($1,$2,$3,$4);",
		"SELECT EndIndex,Data FROM timebatchtable WHERE Key=$1 AND EndTime > $2 ORDER BY EndTime ASC;",
		"SELECT EndIndex,Data FROM timebatchtable WHERE Key=$1 AND EndIndex > $2 ORDER BY EndIndex ASC;",
		"SELECT COALESCE(MAX(EndIndex),0) FROM timebatchtable WHERE Key=$1;",
		"DELETE FROM timebatchtable WHERE Key=$1;")
}

//Uses the correct initializer for the given database driver. The err parameter allows daisychains of errors
func OpenSqlStore(db *sql.DB, sqldriver string, err error) (*SqlStore, error) {
	if err != nil {
		return nil, err
	}
	switch sqldriver {
	case "sqlite3":
		return OpenSQLiteStore(db)
	case "postgres":
		return OpenPostgresStore(db)
	}
	return nil, ERROR_DATABASE_DRIVER
}
