package timeseries

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/karrick/tparse/v2"
)

var SQLVersion = 1

const sqlSchema = `

CREATE TABLE timeseries (
	tsid VARCHAR(36) NOT NULL,
	timestamp REAL NOT NULL,
	duration REAL NOT NULL DEFAULT 0,
	data BLOB,

	PRIMARY KEY (tsid,timestamp),
	CONSTRAINT valid_data CHECK (json_valid(data)),

	CONSTRAINT object_fk
		FOREIGN KEY(tsid)
		REFERENCES objects(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE

);

CREATE TABLE timeseries_actions (
	tsid VARCHAR(36) NOT NULL,
	timestamp REAL NOT NULL,
	duration REAL NOT NULL DEFAULT 0,
	actor VARCHAR DEFAULT NULL,
	data BLOB,

	PRIMARY KEY (tsid,timestamp),
	CONSTRAINT valid_data CHECK (json_valid(data)),

	CONSTRAINT object_fk
		FOREIGN KEY(tsid)
		REFERENCES objects(id)
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
		err = s.rows.Scan(&dp.Timestamp, &dp.Duration, &dp.Actor, &b)
	} else {
		err = s.rows.Scan(&dp.Timestamp, &dp.Duration, &b)
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
		return fmt.Sprintf("(SELECT MIN(timestamp) FROM %s WHERE tsid=?)", table), []interface{}{sid}
	}
	if index == -1 {
		return fmt.Sprintf("(SELECT MAX(timestamp) FROM %s WHERE tsid=?)", table), []interface{}{sid}
	}
	if index > 0 {
		return fmt.Sprintf(`COALESCE(
				(SELECT timestamp FROM %s WHERE tsid=? ORDER BY timestamp ASC LIMIT 1 OFFSET ?),
				(SELECT MAX(timestamp)+1 FROM %s WHERE tsid=?)
			)`, table, table), []interface{}{sid, index, sid}
	}
	// index is < 0, we want to get from the most recent datapoint
	return fmt.Sprintf(`COALESCE(
			(SELECT timestamp FROM %s WHERE tsid=? ORDER BY timestamp DESC LIMIT 1 OFFSET ?),
			(SELECT MIN(timestamp)-1 FROM %s WHERE tsid=?)
		)`, table, table), []interface{}{sid, -index - 1, sid}
}

// generates a query for the given timeseries id. It has all the contents after the "SELECT * FROM " in a query,
// so the result is to be simply pasted instead of manually choosing a table and WHERE clause
// For example, if Query.T is set, will return "timeseries WHERE timestamp=? ORDER BY timestamp ASC", with the timestamp
// in the corresponding value array.
func querySQL(sid string, q *Query, order bool) (string, []interface{}, error) {
	table := "timeseries"
	asc := "ASC"
	constraints := []string{"tsid=?"}
	cValues := []interface{}{sid}

	if q.Actions != nil && *q.Actions {
		table = "timeseries_actions"
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
		t, err := tparse.ParseNow(time.RFC3339, *q.T)
		if err != nil {
			return "", nil, err
		}
		constraints = append(constraints, "timestamp=?")
		cValues = append(cValues, Unix(t))
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
			t, err := tparse.ParseNow(time.RFC3339, *q.T1)
			if err != nil {
				return "", nil, err
			}
			constraints = append(constraints, "timestamp>=?")
			cValues = append(cValues, Unix(t))
		}
		if q.T2 != nil {
			t, err := tparse.ParseNow(time.RFC3339, *q.T2)
			if err != nil {
				return "", nil, err
			}
			constraints = append(constraints, "timestamp<?")
			cValues = append(cValues, Unix(t))
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
	db *database.AdminDB
}

func OpenSQLData(db *database.AdminDB) *SQLData {
	return &SQLData{db: db}
}

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, i *run.Info, curversion int) error {
	if curversion == SQLVersion {
		return nil
	}
	if curversion != 0 {
		return errors.New("Timeseries database version too new")
	}
	_, err := db.ExecUncached(sqlSchema)
	return err
}

func (d *SQLData) TimeseriesDataLength(sid string, actions bool) (l uint64, err error) {
	if actions {
		err = d.db.Get(&l, `SELECT COUNT(*) FROM timeseries_actions WHERE tsid=?`, sid)
		return
	}
	err = d.db.Get(&l, `SELECT COUNT(*) FROM timeseries WHERE tsid=?`, sid)
	return
}

func (d *SQLData) WriteTimeseriesData(sid string, data DatapointIterator, q *InsertQuery) (*Datapoint, float64, float64, int64, error) {
	table := "timeseries"
	insert := "INSERT OR REPLACE"
	ts := float64(-999999999)
	actions := false

	if q.Actions != nil && *q.Actions {
		table = "timeseries_actions"
		actions = true
	}

	dp, err := data.Next()
	if err != nil || dp == nil {
		return dp, ts, ts, 0, err
	}
	tstart := dp.Timestamp
	tend := dp.Timestamp
	count := int64(0)

	tx, err := d.db.Beginx()
	if err != nil {
		return dp, tstart, tend, count, err
	}

	if q.Method != nil && *q.Method != "update" {
		if *q.Method == "append" {
			err = tx.Get(&ts, fmt.Sprintf("SELECT MAX(timestamp) FROM %s WHERE tsid=?", table), sid)
			if err != nil {
				if err != sql.ErrNoRows {
					tx.Rollback()
					return dp, tstart, tend, count, err
				}
				ts = 0
			} else {
			}
		}
		if *q.Method != "insert" {
			return dp, tstart, tend, count, errors.New("Unrecognized insert type")
		}
		insert = "INSERT"
	}
	fullQuery := fmt.Sprintf("%s INTO %s VALUES (?,?,?,?)", insert, table)
	if actions {
		fullQuery = fmt.Sprintf("%s INTO %s VALUES (?,?,?,?,?)", insert, table)
	}
	dp2 := dp
	for dp != nil {
		count++
		dp2 = dp

		if dp.Timestamp <= ts {
			tx.Rollback()
			return dp, tstart, tend, count, errors.New("bad_query: datapoint older than existing data")
		}
		// github.com/vmihailenco/msgpack
		// b, err := msgpack.Marshal(dp.Data)
		b, err := json.Marshal(dp.Data)
		if err != nil {
			tx.Rollback()
			return dp, tstart, tend, count, err
		}
		if actions {
			_, err = tx.Exec(fullQuery, sid, dp.Timestamp, dp.Duration, dp.Actor, b)
		} else {
			_, err = tx.Exec(fullQuery, sid, dp.Timestamp, dp.Duration, b)
		}

		if err != nil {
			tx.Rollback()
			return dp, tstart, tend, count, err
		}
		tend := dp.Timestamp

		dp, err = data.Next()
		if err != nil {
			tx.Rollback()
			return dp, tstart, tend, count, err
		}

	}

	return dp2, tstart, tend, count, tx.Commit()

}

func (d *SQLData) ReadTimeseriesData(sid string, q *Query) (DatapointIterator, error) {

	query, values, err := querySQL(sid, q, true)
	if err != nil {
		return nil, err
	}
	if q.Actions != nil && *q.Actions {
		rows, err := d.db.Queryx("SELECT timestamp,duration,actor,data FROM "+query, values...)

		// TODO: Add transform
		return &SQLIterator{rows.Rows, true}, err
	}
	rows, err := d.db.Queryx("SELECT timestamp,duration,data FROM "+query, values...)

	// TODO: Add transform
	return &SQLIterator{rows.Rows, false}, err

}

func (d *SQLData) RemoveTimeseriesData(sid string, q *Query) error {
	query, values, err := querySQL(sid, q, false)
	if err != nil {
		return err
	}
	_, err = d.db.Exec("DELETE FROM "+query, values...)
	return err

}
