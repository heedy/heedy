package streams

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/vmihailenco/msgpack"
)

const sqlSchema = `

CREATE TABLE streamdata (
	streamid VARCHAR(36),
	timestamp REAL,
	actor VARCHAR DEFAULT NULL,
	data BLOB,

	PRIMARY KEY (streamid,timestamp),

	CONSTRAINT streamfk
		FOREIGN KEY(streamid)
		REFERENCES streams(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE streamdata_actions (
	streamid VARCHAR(36),
	timestamp REAL,
	actor VARCHAR DEFAULT NULL,
	data BLOB,

	PRIMARY KEY (streamid,timestamp),

	CONSTRAINT streamfk
		FOREIGN KEY(streamid)
		REFERENCES streams(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

`

type SQLIterator struct {
	rows *sql.Rows
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

	if err := s.rows.Scan(&dp.Timestamp, &dp.Actor, &b); err != nil {
		s.rows.Close()
		return nil, err
	}
	err := msgpack.Unmarshal(b, &dp.Data)
	return dp, err
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

func (d *SQLData) Length(sid string, actions bool) (l uint64, err error) {
	if actions {
		err = d.db.Get(&l, `SELECT COUNT(*) FROM streamdata_actions WHERE streamid=?`, sid)
		return
	}
	err = d.db.Get(&l, `SELECT COUNT(*) FROM streamdata WHERE streamid=?`, sid)
	return
}

func (d *SQLData) Insert(sid string, data DatapointIterator, q *InsertQuery) error {
	table := "streamdata"
	insert := "INSERT"
	ts := float64(-999999999)

	if q.Actions != nil && *q.Actions {
		table = "streamdata_actions"
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
	fullQuery := fmt.Sprintf("%s INTO %s VALUES (?,?,?,?)", insert, table)

	for dp != nil {
		if dp.Timestamp <= ts {
			tx.Rollback()
			return errors.New("bad_query: datapoint older than existing data")
		}
		b, err := msgpack.Marshal(dp.Data)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.Exec(fullQuery, sid, dp.Timestamp, dp.Actor, b)
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

func (d *SQLData) Read(sid string, q *Query) (*SQLIterator, error) {
	table := "streamdata"
	asc := "ASC"

	if q.Actions != nil && *q.Actions {
		table = "streamdata_actions"
	}
	if q.Reversed != nil && *q.Reversed {
		asc = "DESC"
	}
	constraints := []string{"streamid=?"}
	cValues := []interface{}{sid}

	if q.T1 != nil {
		constraints = append(constraints, "timestamp>=?")
		cValues = append(cValues, *q.T1)
	}
	if q.T2 != nil {
		constraints = append(constraints, "timestamp<?")
		cValues = append(cValues, *q.T2)
	}
	if q.T != nil {
		if len(constraints) > 1 {
			return nil, errors.New("bad_query: can't query both by time range and by timestamp at the same time")
		}
		constraints = append(constraints, "timestamp=?")
		cValues = append(cValues, *q.T)
	}
	if q.I == nil && q.I1 == nil && q.I2 == nil {
		rows, err := d.db.Query(fmt.Sprintf("SELECT timestamp,actor,data FROM %s WHERE %s ORDER BY timestamp %s;", table, strings.Join(constraints, " AND "), asc), cValues...)
		return &SQLIterator{rows}, err
	}

	if len(constraints) > 1 {
		return nil, errors.New("bad_query: can't query by time and by index at the same time")
	}

	return nil, errors.New("server_error: querying by index is currently not supported")
}

func (d *SQLData) Remove(sid string, q *Query) error {
	table := "streamdata"
	if q.Actions != nil && *q.Actions {
		table = "streamdata_actions"
	}
	constraints := []string{"streamid=?"}
	cValues := []interface{}{sid}

	if q.T1 != nil {
		constraints = append(constraints, "timestamp>=?")
		cValues = append(cValues, *q.T1)
	}
	if q.T2 != nil {
		constraints = append(constraints, "timestamp<?")
		cValues = append(cValues, *q.T2)
	}
	if q.T != nil {
		if len(constraints) > 1 {
			return errors.New("bad_query: can't remove both by time range and by timestamp at the same time")
		}
		constraints = append(constraints, "timestamp=?")
		cValues = append(cValues, *q.T)
	}
	if q.I == nil && q.I1 == nil && q.I2 == nil {
		_, err := d.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s;", table, strings.Join(constraints, " AND ")), cValues...)
		return err
	}

	if len(constraints) > 1 {
		return errors.New("bad_query: can't remove by time and by index at the same time")
	}

	return errors.New("server_error: querying by index is currently not supported")
}
