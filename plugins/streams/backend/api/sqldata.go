package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/heedy/heedy/backend/database"
)

var SQLVersion = 1

const sqlSchema = `

CREATE TABLE streamdata (
	streamid VARCHAR(36),
	timestamp REAL,
	data BLOB,

	PRIMARY KEY (streamid,timestamp),
	CONSTRAINT valid_data CHECK (json_valid(data)),

	CONSTRAINT source_fk
		FOREIGN KEY(streamid)
		REFERENCES sources(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE

);

CREATE TABLE streamdata_actions (
	streamid VARCHAR(36),
	timestamp REAL,
	actor VARCHAR DEFAULT NULL,
	data BLOB,

	PRIMARY KEY (streamid,timestamp),
	CONSTRAINT valid_data CHECK (json_valid(data)),

	CONSTRAINT source_fk
		FOREIGN KEY(streamid)
		REFERENCES sources(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

`

type SQLIterator struct {
	rows    *sql.Rows
	actions bool
}

func (s *SQLIterator) Close() error {
	return s.rows.Close()
}

func (s *SQLIterator) Next() (*Datapoint, error) {
	n := s.rows.Next()
	if !n {
		s.rows.Close()
		return nil, nil
	}
	dp := &Datapoint{}
	var b []byte

	var err error
	if s.actions {
		err = s.rows.Scan(&dp.Timestamp, &dp.Actor, &b)
	} else {
		err = s.rows.Scan(&dp.Timestamp, &b)
	}

	if err != nil {
		s.rows.Close()
		return nil, err
	}
	// github.com/vmihailenco/msgpack
	// err := msgpack.Unmarshal(b, &dp.Data)
	err = json.Unmarshal(b, &dp.Data)
	return dp, err
}

// returns the timestamp associated with the given index. If the index goes beyond the array bounds, returns a timestamp
// 1 beyond the earliest/latest datapoint. This allows fuzzy index-based querying.
func getSQLIndexTimestamp(table string, sid string, index int64) (string, []interface{}) {
	if index == 0 {
		return fmt.Sprintf("(SELECT MIN(timestamp) FROM %s WHERE streamid=?)", table), []interface{}{sid}
	}
	if index == -1 {
		return fmt.Sprintf("(SELECT MAX(timestamp) FROM %s WHERE streamid=?)", table), []interface{}{sid}
	}
	if index > 0 {
		return fmt.Sprintf(`COALESCE(
				(SELECT timestamp FROM %s WHERE streamid=? ORDER BY timestamp ASC LIMIT 1 OFFSET ?),
				(SELECT MAX(timestamp)+1 FROM %s WHERE streamid=?)
			)`, table, table), []interface{}{sid, index, sid}
	}
	// index is < 0, we want to get from the most recent datapoint
	return fmt.Sprintf(`COALESCE(
			(SELECT timestamp FROM %s WHERE streamid=? ORDER BY timestamp DESC LIMIT 1 OFFSET ?),
			(SELECT MIN(timestamp)-1 FROM %s WHERE streamid=?)
		)`, table, table), []interface{}{sid, -index - 1, sid}
}

// generates a query for the given stream id. It has all the contents after the "SELECT * FROM " in a query,
// so the result is to be simply pasted instead of manually choosing a table and WHERE clause
// For example, if Query.T is set, will return "streamdata WHERE timestamp=? ORDER BY timestamp ASC", with the timestamp
// in the corresponding value array.
func querySQL(sid string, q *Query, order bool) (string, []interface{}, error) {
	table := "streamdata"
	asc := "ASC"
	constraints := []string{"streamid=?"}
	cValues := []interface{}{sid}

	if q.Actions != nil && *q.Actions {
		table = "streamdata_actions"
	}
	if q.Reversed != nil {
		if *q.Reversed {
			asc = "DESC"
		}
		if order {
			return "", nil, errors.New("bad_query: Ordering is not supported on this query type")
		}
	}

	if q.T != nil {
		constraints = append(constraints, "timestamp=?")
		cValues = append(cValues, *q.T)
		if q.I != nil || q.I1 != nil || q.I2 != nil || q.T1 != nil || q.T2 != nil {
			return "", nil, errors.New("bad_query: Cannot query by range and by single timestamp at the same time")
		}
	} else if q.I != nil {
		c, v := getSQLIndexTimestamp(table, sid, *q.I)
		constraints = append(constraints, "timestamp="+c)
		cValues = append(cValues, v...)
		if q.I1 != nil || q.I2 != nil || q.T1 != nil || q.T2 != nil {
			return "", nil, errors.New("bad_query: Cannot query by range and by single index at the same time")
		}
	} else {
		// Otherwise, we're querying a range
		if q.T1 != nil {
			constraints = append(constraints, "timestamp>=?")
			cValues = append(cValues, *q.T1)
		}
		if q.T2 != nil {
			constraints = append(constraints, "timestamp<?")
			cValues = append(cValues, *q.T2)
		}
		if q.I1 != nil {
			c, v := getSQLIndexTimestamp(table, sid, *q.I1)
			constraints = append(constraints, "timestamp>="+c)
			cValues = append(cValues, v...)
		}
		if q.I2 != nil {
			c, v := getSQLIndexTimestamp(table, sid, *q.I2)
			constraints = append(constraints, "timestamp<"+c)
			cValues = append(cValues, v...)
		}

	}

	// If ordering is not supported, return a query without the order by clause
	if !order {
		if q.Transform != nil || q.Limit != nil {
			return "", nil, errors.New("bad_query: limits and transforms are not supported on this query type")
		}
		return fmt.Sprintf("%s WHERE %s", table, strings.Join(constraints, " AND ")), cValues, nil
	}

	totalQuery := fmt.Sprintf("%s WHERE %s ORDER BY timestamp %s", table, strings.Join(constraints, " AND "), asc)
	if q.Transform == nil && q.Limit != nil {
		totalQuery = totalQuery + " LIMIT ?"
		cValues = append(cValues, *q.Limit)
	}
	return totalQuery, cValues, nil
}

type SQLData struct {
	db *sqlx.DB
}

func CreateSQLData(db *sqlx.DB) error {
	_, err := db.Exec(sqlSchema)
	return err
}

func OpenSQLData(db *sqlx.DB) *SQLData {
	return &SQLData{db: db}
}

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, curversion int) error {
	if curversion != 0 {
		return errors.New("Streams database version too new")
	}
	return CreateSQLData(db.DB)
}

func (d *SQLData) StreamDataLength(sid string, actions bool) (l uint64, err error) {
	if actions {
		err = d.db.Get(&l, `SELECT COUNT(*) FROM streamdata_actions WHERE streamid=?`, sid)
		return
	}
	err = d.db.Get(&l, `SELECT COUNT(*) FROM streamdata WHERE streamid=?`, sid)
	return
}

func (d *SQLData) WriteStreamData(sid string, data DatapointIterator, q *InsertQuery) error {
	table := "streamdata"
	insert := "INSERT"
	ts := float64(-999999999)
	actions := false

	if q.Actions != nil && *q.Actions {
		table = "streamdata_actions"
		actions = true
	}

	dp, err := data.Next()
	if err != nil || dp == nil {
		return err
	}

	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}

	if q.Type != nil && *q.Type != "INSERT" {
		if *q.Type == "append" {
			err = tx.Get(&ts, fmt.Sprintf("SELECT MAX(timestamp) FROM %s WHERE streamid=?", table), sid)
			if err != nil {
				if err != sql.ErrNoRows {
					tx.Rollback()
					return err
				}
				ts = 0
			} else {
			}
		}
		insert = "INSERT OR REPLACE"
	}
	fullQuery := fmt.Sprintf("%s INTO %s VALUES (?,?,?)", insert, table)
	if actions {
		fullQuery = fmt.Sprintf("%s INTO %s VALUES (?,?,?,?)", insert, table)
	}

	for dp != nil {
		if dp.Timestamp <= ts {
			tx.Rollback()
			return errors.New("bad_query: datapoint older than existing data")
		}
		// github.com/vmihailenco/msgpack
		// b, err := msgpack.Marshal(dp.Data)
		b, err := json.Marshal(dp.Data)
		if err != nil {
			tx.Rollback()
			return err
		}
		if actions {
			_, err = tx.Exec(fullQuery, sid, dp.Timestamp, dp.Actor, b)
		} else {
			_, err = tx.Exec(fullQuery, sid, dp.Timestamp, b)
		}

		if err != nil {
			tx.Rollback()
			return err
		}

		dp, err = data.Next()
		if err != nil {
			tx.Rollback()
			return err
		}

	}

	return tx.Commit()

}

func (d *SQLData) ReadStreamData(sid string, q *Query) (DatapointIterator, error) {

	query, values, err := querySQL(sid, q, true)
	if err != nil {
		return nil, err
	}
	if q.Actions != nil && *q.Actions {
		rows, err := d.db.Query("SELECT timestamp,actor,data FROM "+query, values...)

		// TODO: Add transform
		return &SQLIterator{rows, true}, err
	}
	rows, err := d.db.Query("SELECT timestamp,data FROM "+query, values...)

	// TODO: Add transform
	return &SQLIterator{rows, false}, err

}

func (d *SQLData) RemoveStreamData(sid string, q *Query) error {
	query, values, err := querySQL(sid, q, false)
	if err != nil {
		return err
	}
	_, err = d.db.Exec("DELETE FROM "+query, values...)
	return err

}
