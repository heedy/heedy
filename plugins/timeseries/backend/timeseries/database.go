package timeseries

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/pipescript"
	"github.com/jmoiron/sqlx"
	"github.com/klauspost/compress/zstd"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
)

/* So originally, heedy's timeseries were in the "standard" format, with a row per datapoint.
However, this led to the following problems:
1) Extremely slow inserts - 1 million datapoints took over 40 seconds!
2) Slow reads - a query on 1 million datapoints took 5 seconds, and profiled it was spending a huge amount of time in Scan
3) Very large disk footprint - a database with 1 million datapoints was over 100MB, whereas the pure json array was ~30MB

The slow reads are largely due to problems with golang's Scan being slow (https://github.com/mattn/go-sqlite3/issues/379).
Slow writes were due to a large amount of checks that needed to be done for overlapping datapoints and such. Finally, a lack
of compression support was the culprit of large space usage.

I therefore chose to rewrite the timeseries table by batching elements. This is not a decision I take lightly - it is generally bad practice
to ""roll your own" storage format within SQL, and all interpretability is lost for timeseries rows, so they now look like gibberish, and
they can no longer be analyzed by SQL json primitives. Not to mention how difficult and bug-prone reading/writing this custom format can be.

On the other hand, the benefits are fixes to the above 3 issues. Reads are now extremely fast, each batch is compressed, lowering database size,
and large inserts happen much faster. The usage of the timeseries in heedy works well with a batching mechanism, since PipeScript is used to filter
the data rather than raw SQL. I therefore deemed it more important that heedy can query and add to timeseries extremely fast than the simplicity
of the direct approach.

*/

var SQLVersion = 1

// sqlSchema is initialized in plugin.go (SQLUpdater)
const sqlSchema = `

CREATE TABLE timeseries (
	tsid VARCHAR(36) NOT NULL,
	tstart REAL NOT NULL,
	tend REAL NOT NULL,
	length INTEGER NOT NULL,

	-- timeseries data comes as zstandard-compressed msgpack array batches
	data BLOB,

	PRIMARY KEY (tsid,tstart),
	CONSTRAINT valid_range CHECK (tstart <= tend AND length > 0),

	CONSTRAINT object_fk
		FOREIGN KEY(tsid)
		REFERENCES objects(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE

);
CREATE INDEX timeseries_duration ON timeseries(tsid,tend,tstart);
`

/* Timeseries actions are not enabled for 0.4
`
CREATE TABLE timeseries_actions (
	tsid VARCHAR(36) NOT NULL,
	tstart REAL NOT NULL,
	tend REAL NOT NULL,
	length INTEGER NOT NULL,
	data BLOB,

	PRIMARY KEY (tsid,tstart),
	CONSTRAINT valid_range CHECK (tstart <= tend AND length > 0),


	CONSTRAINT object_fk
		FOREIGN KEY(tsid)
		REFERENCES objects(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);
CREATE INDEX timeseries_actions_duration ON timeseries_actions(tsid,tend,tstart);
`
*/

//go:generate msgp -o=database_msgp.go -tests=false
//msgp:ignore Query
//msgp:ignore TimeseriesDB

//easyjson:json
type Datapoint struct {
	Timestamp float64     `json:"t" db:"timestamp" msg:"t"`
	Duration  float64     `json:"dt,omitempty" db:"duration" msg:"dt,omitempty"`
	Data      interface{} `json:"d" db:"data" msg:"d"`

	Actor string `json:"a,omitempty" db:"actor" msg:"a,omitempty"`
}

//IsEqual checks if the datapoint is equal to another datapoint
func (d *Datapoint) IsEqual(dp *Datapoint) bool {
	return (dp.Timestamp == d.Timestamp && dp.Duration == d.Duration && dp.Actor == d.Actor && pipescript.Equal(d.Data, dp.Data))
}

func (d *Datapoint) EndTime() float64 {
	return d.Timestamp + d.Duration
}

// String returns a json representation of the datapoint
func (d *Datapoint) String() string {
	b, _ := easyjson.Marshal(d)
	return string(b)
}

// NewDatapoint returns a datapoint with the current timestamp
func NewDatapoint(data interface{}) *Datapoint {
	return &Datapoint{
		Timestamp: float64(time.Now().UnixNano()) * 1e-9,
		Data:      data,
	}
}

//A DatapointArray holds a couple useful functions that act on it
//easyjson:json
type DatapointArray []*Datapoint

// String returns a json representation of the datapoint
func (dpa DatapointArray) String() string {
	b, _ := easyjson.Marshal(dpa)
	return string(b)
}

//IsEqual checks if two DatapointArrays contain the same data
func (dpa DatapointArray) IsEqual(d DatapointArray) bool {
	if len(d) != len(dpa) {
		return false
	}
	for i := range d {
		if !d[i].IsEqual(dpa[i]) {
			return false
		}
	}
	return true
}

// The zstandard encoder used for compression into the database
var zencoder *zstd.Encoder
var zdecoder, _ = zstd.NewReader(nil)

func (dpa DatapointArray) ToBytes() ([]byte, error) {
	//b, err := easyjson.Marshal(dpa)
	b, err := dpa.MarshalMsg(nil)

	if err != nil || zencoder == nil {
		return b, err
	}

	b = zencoder.EncodeAll(b, make([]byte, 0, len(b)/2))
	return b, err
}

//DatapointArrayFromBytes decompresses a gzipped byte array for the compressed representation of a DatapointArray
func DatapointArrayFromBytes(b []byte) (dpa DatapointArray, err error) {
	if zencoder != nil {
		b, err = zdecoder.DecodeAll(b, make([]byte, 0, len(b)*10))
		if err != nil {
			return nil, err
		}
	}
	//easyjson.Unmarshal(b, &dpa)
	_, err = dpa.UnmarshalMsg(b)
	return
}

type TimeseriesDB struct {
	DB                    *database.AdminDB `mapstructure:"-"`
	BatchSize             int               `mapstructure:"batch_size"`
	MaxBatchSize          int               `mapstructure:"max_batch_size"`
	BatchCompressionLevel int               `mapstructure:"batch_compression_level"`
	CompressQueryResponse bool              `mapstructure:"compress_query_response"`
}

func (ts *TimeseriesDB) Length(tsid string, actions bool) (l int64, err error) {
	table := "timeseries"
	if actions {
		table = "timeseries_actions"
	}

	err = ts.DB.Get(&l, fmt.Sprintf("SELECT COALESCE(SUM(length),0) FROM %s WHERE tsid=?", table), tsid)
	return
}

type Query struct {
	Timeseries string      `json:"timeseries,omitempty"`
	T1         interface{} `json:"t1,omitempty"`
	T2         interface{} `json:"t2,omitempty"`
	I1         *int64      `json:"i1,omitempty" schema:"i1"`
	I2         *int64      `json:"i2,omitempty" schema:"i2"`
	Limit      *int64      `json:"limit,omitempty" schema:"limit"`
	T          interface{} `json:"t,omitempty"`
	I          *int64      `json:"i,omitempty" schema:"i"`
	Transform  *string     `json:"transform,omitempty" schema:"transform"`
	Actions    *bool       `json:"actions,omitempty" schema:"actions"`
}

// String returns a json representation of the datapoint
func (q Query) String() string {
	b, _ := json.Marshal(q)
	return string(b)
}

func (ts *TimeseriesDB) rawQuery(q *Query) (DatapointIterator, error) {
	table := "timeseries"
	if q.Timeseries == "" {
		return nil, errors.New("bad_query: no timeseries specified")
	}

	if q.Actions != nil && *q.Actions {
		table = "timeseries_actions"
	}
	constraints := []string{"tsid=?"}
	cValues := []interface{}{q.Timeseries}

	var err error
	var t1, t2 float64

	// The timestamps are parsed here, because they are used in both time range and index queries
	if q.T1 != nil {
		t1, err = ParseTimestamp(q.T1)
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, "tend >= ?")
		cValues = append(cValues, t1)
	}
	if q.T2 != nil {
		t2, err = ParseTimestamp(q.T2)
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, "tstart <= ?")
		cValues = append(cValues, t2)
	}

	// If query by time only
	if q.I == nil && (q.I1 == nil || *q.I1 == 0) && q.I2 == nil {
		// This is a query that only uses time ranges. It means that we don't need to worry about any index values,
		// and don't need to set up any complex transaction logic

		if q.T != nil {
			// Return only the datapoint at this specific time
			if q.T1 != nil || q.T2 != nil {
				return nil, errors.New("bad_query: Cannot query by range and by single timestamp at the same time")
			}
			t, err := ParseTimestamp(q.T)
			if err != nil {
				return nil, err
			}
			rows, err := ts.DB.Queryx(fmt.Sprintf("SELECT data FROM %s WHERE tsid=? AND tstart<=? AND tend >=? ORDER BY tstart ASC", table), q.Timeseries, t, t)
			if err != nil {
				return nil, err
			}
			bi := SQLBatchIterator{rows, nil}
			defer bi.Close()
			da, err := bi.NextBatch()
			if err != nil || da == nil {
				return EmptyIterator{}, err
			}
			da, err = BatchTOffset(bi, da, t)
			if err != nil || da == nil {
				return EmptyIterator{}, err
			}
			if da[0].Timestamp != t {
				return EmptyIterator{}, err
			}
			return NewDatapointArrayIterator(da[:1]), nil
		}

		rows, err := ts.DB.Queryx(fmt.Sprintf("SELECT data FROM %s WHERE %s ORDER BY tstart ASC", table, strings.Join(constraints, " AND ")), cValues...)
		if err != nil {
			return nil, err
		}
		bi := BatchIterator(SQLBatchIterator{rows, nil})
		da, err := bi.NextBatch()
		if err != nil || da == nil {
			return EmptyIterator{}, err
		}
		if q.T1 != nil {
			da, err = BatchTOffset(bi, da, t1)
			if err != nil || len(da) == 0 {
				return EmptyIterator{}, err
			}
		}
		if q.T2 != nil {
			bi = BatchEndTime{bi, t2}
		}
		return NewBatchDatapointIterator(NewChanBatchIterator(bi), da), nil

	}
	if q.T != nil {
		return nil, errors.New("bad_query: cannot query by single index and single timestamp at the same time")
	}

	// There is at least one index-based query. This sucks, since it means we have to do extra querying to figure out where exactly the index is
	if q.I != nil {
		// Return only the datapoint at the specific index
		if q.T1 != nil || q.T2 != nil || q.I1 != nil || q.I2 != nil {
			return nil, errors.New("bad_query: Cannot query by range and by single index at the same time")
		}
		// The batch timestamp and offset from beginning of batch of the index
		var startOffset struct {
			Data   []byte
			Offset int
		}
		if *q.I >= 0 {
			// The index is going to be computed from the start of the data
			err = ts.DB.Get(&startOffset, fmt.Sprintf(`WITH q AS (SELECT tstart,length,data,SUM(length) OVER (ORDER BY tstart ASC) AS endindex FROM %s WHERE tsid=?) SELECT data,?-(endindex-length) AS offset FROM q WHERE endindex-length<=? AND endindex>? LIMIT 1;`, table), q.Timeseries, *q.I, *q.I, *q.I)
		} else {
			// The index is negative, so compute from end
			err = ts.DB.Get(&startOffset, fmt.Sprintf(`WITH q AS (SELECT tstart,length,data,SUM(length) OVER (ORDER BY tstart DESC) AS negindex FROM %s WHERE tsid=?) SELECT data,negindex-? AS offset FROM q WHERE negindex>=? AND negindex-length<? LIMIT 1;`, table), q.Timeseries, -*q.I, -*q.I, -*q.I)
		}
		if err != nil {
			if err == sql.ErrNoRows {
				return EmptyIterator{}, nil
			}
			return nil, err
		}
		da, err := DatapointArrayFromBytes(startOffset.Data)
		if err != nil {
			return nil, err
		}
		return NewDatapointArrayIterator(da[startOffset.Offset : startOffset.Offset+1]), nil
	}

	// At this point, we have some kind of range that includes an index. We unfortunately need to do this in a transaction, since
	// want to actually get correct indices
	tx, err := ts.DB.Beginx()
	if err != nil {
		return nil, err
	}

	var so, so2 struct {
		Tstart float64
		Offset int
	}
	readlimit := int64(0) // Limit the number of datapoints to read

	i1 := q.I1
	i2 := q.I2

	if i1 != nil {
		if *q.I1 >= 0 {
			err = tx.Get(&so, fmt.Sprintf(`WITH q AS (SELECT tstart,length,SUM(length) OVER (ORDER BY tstart ASC) AS endindex FROM %s WHERE tsid=?) SELECT tstart,?-(endindex-length) AS offset FROM q WHERE endindex-length<=? AND endindex>? LIMIT 1;`, table), q.Timeseries, *i1, *i1, *i1)
			if err == sql.ErrNoRows {
				// I1 goes beyond range - return empty array
				tx.Rollback()
				return EmptyIterator{}, nil
			}
		} else {
			err = tx.Get(&so, fmt.Sprintf(`WITH q AS (SELECT tstart,length,SUM(length) OVER (ORDER BY tstart DESC) AS negindex FROM %s WHERE tsid=?) SELECT tstart,negindex-? AS offset FROM q WHERE negindex>=? AND negindex-length<? LIMIT 1;`, table), q.Timeseries, -*i1, -*i1, -*i1)
			if err == sql.ErrNoRows {
				// I1 reverses beyond range - so basically start at 0, and ignore the index
				i1 = nil
				err = nil
			}
		}
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if i1 != nil {
			constraints = append(constraints, "tstart >= ?")
			cValues = append(cValues, so.Tstart)
		}

	}
	if q.I2 != nil {
		if i1 != nil && (q.T1 == nil || t1 <= so.Tstart) && (*i1 < 0 && *i2 < 0 || *i1 >= 0 && *i2 >= 0) {
			// We actually don't need to find the end index, because we can encode the query as a limit reader
			readlimit = *i2 - *i1
			if readlimit <= 0 {
				tx.Rollback()
				// The index is more negative
				return EmptyIterator{}, nil
			}
			i2 = nil // Set the end index to empty
		} else {
			if *q.I2 >= 0 {
				err = tx.Get(&so2, fmt.Sprintf(`WITH q AS (SELECT tstart,length,SUM(length) OVER (ORDER BY tstart ASC) AS endindex FROM %s WHERE tsid=?) SELECT tstart,?-(endindex-length) AS offset FROM q WHERE endindex-length<=? AND endindex>? LIMIT 1;`, table), q.Timeseries, *i2, *i2, *i2)
				if err == sql.ErrNoRows {
					// I2 goes beyond range - so basically eliminate the end constraint
					i2 = nil
					err = nil
				}
			} else {
				err = tx.Get(&so2, fmt.Sprintf(`WITH q AS (SELECT tstart,length,SUM(length) OVER (ORDER BY tstart DESC) AS negindex FROM %s WHERE tsid=?) SELECT tstart,negindex-? AS offset FROM q WHERE negindex>=? AND negindex-length<? LIMIT 1;`, table), q.Timeseries, -*i2, -*i2, -*i2)
				if err == sql.ErrNoRows {
					// I2 reverses beyond range - return empty array
					tx.Rollback()
					return EmptyIterator{}, nil
				}
			}
		}
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if i2 != nil {
			constraints = append(constraints, "tstart <= ?")
			cValues = append(cValues, so2.Tstart)
		}
	}

	// Now query the full thing

	rows, err := tx.Queryx(fmt.Sprintf("SELECT data FROM %s WHERE %s ORDER BY tstart ASC", table, strings.Join(constraints, " AND ")), cValues...)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	bi := BatchIterator(SQLBatchIterator{rows, func() { tx.Commit() }})
	if i2 != nil {
		bi = BatchEndOffset{bi, so2.Tstart, so2.Offset}
	}
	if q.T2 != nil {
		bi = BatchEndTime{bi, t2}
	}
	da, err := bi.NextBatch()
	if err != nil || da == nil {
		bi.Close()
		return EmptyIterator{}, err
	}
	if i1 != nil && so.Tstart == da[0].Timestamp {
		if len(da) <= so.Offset {
			bi.Close()
			return EmptyIterator{}, nil
		}
		da = da[so.Offset:]
	}
	if readlimit > 0 {
		if readlimit < int64(len(da)) {
			da = da[:readlimit]
			readlimit = 0
		} else {
			readlimit -= int64(len(da))
		}
		bi = &BatchPointLimit{bi, readlimit}
	}
	if q.T1 != nil {
		da, err = BatchTOffset(bi, da, t1)
		if err != nil || len(da) == 0 {
			bi.Close()
			return EmptyIterator{}, err
		}
	}

	return NewBatchDatapointIterator(NewChanBatchIterator(bi), da), nil
}

// Query runs the given query, while adding on the transform and limit reading
func (ts *TimeseriesDB) Query(q *Query) (DatapointIterator, error) {
	it, err := ts.rawQuery(q)
	if err != nil {
		return it, err
	}
	if q.Transform != nil && *q.Transform != "" {
		it2 := it
		it, err = NewTransformIterator(*q.Transform, it2)
		if err != nil {
			it2.Close()
		}
	}
	if q.Limit != nil && *q.Limit > 0 {
		it = NewNumIterator(it, *q.Limit)
	}

	return it, err
}

func (ts *TimeseriesDB) Delete(q *Query) error {
	table := "timeseries"

	if q.Timeseries == "" {
		return errors.New("bad_query: no timeseries specified")
	}

	if q.Actions != nil && *q.Actions {
		table = "timeseries_actions"
	}

	var err error
	t1 := math.Inf(-1)
	t2 := math.Inf(1)

	var so1, so2 struct {
		Tstart float64
		Offset int
	}
	i1 := q.I1
	i2 := q.I2

	if q.Transform != nil || q.Limit != nil {
		return errors.New("bad_query: transforms and limits are not supported for delete")
	}

	if q.T1 != nil {
		t1, err = ParseTimestamp(q.T1)
		if err != nil {
			return err
		}
	}
	if q.T2 != nil {
		t2, err = ParseTimestamp(q.T2)
		if err != nil {
			return err
		}
	}

	tx, err := ts.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete first finds the bounds over which to delete, deletes all internal elements in the range,
	// and then finally handles the lower and upper bound batches

	if q.T != nil {
		if q.I != nil || q.T1 != nil || q.T2 != nil || q.I1 != nil || q.I2 != nil {
			return errors.New("bad_query: cannot delete by single timestamp with additional range/index")
		}
		// If T is defined, let both t1 and t2 be T, we special-case the t1=t2 situation
		t1, err = ParseTimestamp(q.T)
		if err != nil {
			return err
		}
		t2 = t1

	} else {
		if q.I != nil {
			if q.T1 != nil || q.T2 != nil || q.I1 != nil || q.I2 != nil {
				return errors.New("bad_query: cannot delete by single index with additional range")
			}
			i1 = q.I
			if *q.I != -1 {
				i2v := *q.I + 1
				i2 = &i2v
			}

		}
		if i1 != nil {
			if *i1 >= 0 {
				// The index is going to be computed from the start of the data
				err = tx.Get(&so1, fmt.Sprintf(`WITH q AS (SELECT tstart,length,SUM(length) OVER (ORDER BY tstart ASC) AS endindex FROM %s WHERE tsid=?) SELECT tstart,?-(endindex-length) AS offset FROM q WHERE endindex-length<=? AND endindex>? LIMIT 1;`, table), q.Timeseries, *i1, *i1, *i1)
				if err != nil {
					if err == sql.ErrNoRows {
						return nil // If it goes beyond end, nothing to delete
					}
					return err
				}
			} else {
				// The index is negative, so compute from end
				err = tx.Get(&so1, fmt.Sprintf(`WITH q AS (SELECT tstart,length,SUM(length) OVER (ORDER BY tstart DESC) AS negindex FROM %s WHERE tsid=?) SELECT tstart,negindex-? AS offset FROM q WHERE negindex>=? AND negindex-length<? LIMIT 1;`, table), q.Timeseries, -*i1, -*i1, -*i1)
				if err != nil {
					if err != sql.ErrNoRows {
						return err
					}
					i1 = nil // If it goes beyond end, we just delete to the end
				}
			}
		}
		if i2 != nil {
			if *i2 >= 0 {
				// The index is going to be computed from the start of the data
				err = tx.Get(&so2, fmt.Sprintf(`WITH q AS (SELECT tstart,length,SUM(length) OVER (ORDER BY tstart ASC) AS endindex FROM %s WHERE tsid=?) SELECT tstart,?-(endindex-length) AS offset FROM q WHERE endindex-length<=? AND endindex>? LIMIT 1;`, table), q.Timeseries, *i2, *i2, *i2)
				if err != nil {
					if err != sql.ErrNoRows {
						return err
					}
					i2 = nil // If it goes beyond end, we just delete to the end
				}
			} else {
				// The index is negative, so compute from end
				err = tx.Get(&so2, fmt.Sprintf(`WITH q AS (SELECT tstart,length,SUM(length) OVER (ORDER BY tstart DESC) AS negindex FROM %s WHERE tsid=?) SELECT tstart,negindex-? AS offset FROM q WHERE negindex>=? AND negindex-length<? LIMIT 1;`, table), q.Timeseries, -*i2, -*i2, -*i2)
				if err != nil {
					if err == sql.ErrNoRows {
						return nil
					}
					return err
				}
			}
		}
	}

	// OK, all constraints are now encoded in a way we can use. Now, delete all batches that are entirely contained between the start and end of the range
	if q.I == nil && q.T == nil { // single datapoint deletes don't need this delete, since they happen entirely over batches
		constraints := []string{"tsid=?"}
		cValues := []interface{}{q.Timeseries}

		if q.T1 != nil {
			constraints = append(constraints, "tstart >= ?")
			cValues = append(cValues, t1)
		}
		if q.T2 != nil {
			constraints = append(constraints, "tend < ?")
			cValues = append(cValues, t2)
		}
		if i1 != nil {
			constraints = append(constraints, "tstart > ?")
			cValues = append(cValues, so1.Tstart)
		}
		if i2 != nil {
			constraints = append(constraints, "tstart < ?")
			cValues = append(cValues, so2.Tstart)
		}

		_, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s", table, strings.Join(constraints, " AND ")), cValues...)
		if err != nil {
			return err
		}
	}

	// Now we only have the 2 endpoints to worry about, each of which can be up to 2 batches. We take this full range, manually filter it, and
	// then re-add the remaining data
	constraints := []string{"tsid=?"}
	cValues := []interface{}{q.Timeseries}

	if q.T1 != nil || q.T != nil {
		constraints = append(constraints, "tend >= ?")
		cValues = append(cValues, t1)
	}
	if q.T2 != nil || q.T != nil {
		constraints = append(constraints, "tstart <= ?")
		cValues = append(cValues, t2)
	}
	if i1 != nil {
		constraints = append(constraints, "tstart >= ?")
		cValues = append(cValues, so1.Tstart)
	}
	if i2 != nil {
		constraints = append(constraints, "tstart <= ?")
		cValues = append(cValues, so2.Tstart)
	}

	rows, err := tx.Queryx(fmt.Sprintf("SELECT data FROM %s WHERE %s ORDER BY tstart ASC", table, strings.Join(constraints, " AND ")), cValues...)
	if err != nil {
		return err
	}
	bi := SQLBatchIterator{rows, nil}
	dpa := make(DatapointArray, 0, ts.MaxBatchSize)

	i := 0
	for ; i < 5; i++ {
		dp2, err := bi.NextBatch()
		if err != nil {
			return err
		}
		if dp2 == nil {
			break
		}

		// Before adding the batch, check if it is the one referenced by index range, so that we can simplify the filtering to between a time range only
		if i1 != nil && dp2[0].Timestamp == so1.Tstart {
			if dp2[so1.Offset].Timestamp > t1 {
				t1 = dp2[so1.Offset].Timestamp
			}
		}
		if i2 != nil && dp2[0].Timestamp == so2.Tstart {
			if dp2[so2.Offset].Timestamp < t2 {
				t2 = dp2[so2.Offset].Timestamp
			}
		}

		dpa = append(dpa, dp2...)
	}
	if i >= 5 {
		return errors.New("database_corrupted: A database constraint was violated - this is a critical bug")
	}
	bi.Close()

	// Now that the data was loaded, let's delete those batches from the database - we will be replacing them
	_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s", table, strings.Join(constraints, " AND ")), cValues...)
	if err != nil {
		return err
	}

	if len(dpa) > 0 {

		// Next, filter the data!
		ndpa := make(DatapointArray, 0, len(dpa))
		for i = 0; i < len(dpa); i++ {
			if dpa[i].Timestamp < t1 || dpa[i].Timestamp >= t2 {
				if q.T == nil || dpa[i].Timestamp != t1 {
					ndpa = append(ndpa, dpa[i])
				}
			}
		}

		// ndpa now holds all filtered data. Now split it into batches and reinsert into the timeseries
		if len(ndpa) > 0 {
			ndpa, err = ts.writeBatchStart(tx, table, q.Timeseries, ndpa)
			if err != nil {
				return err
			}
			err = ts.writeBatch(tx, table, q.Timeseries, ndpa)
			if err != nil {
				return err
			}
		}

	}

	return tx.Commit()
}

// IteratedBatcher is basically like an SQLBatchIterator, but it closes the sql connection in-between calls to NextBatch,
// so that updates/edits can happen to the database in the mean time.
type IteratedBatcher struct {
	tx        database.TxWrapper
	tstart    float64
	tsid      string
	queueSize int
	done      bool
	table     string

	queuedbatches []DatapointArray
}

func (ib *IteratedBatcher) Close() error {
	return nil // Don't close the transaction, since it is probably being used for other stuff outside here
}

func (ib *IteratedBatcher) NextBatch() (DatapointArray, error) {
	if len(ib.queuedbatches) > 0 {
		batch := ib.queuedbatches[0]
		ib.queuedbatches = ib.queuedbatches[1:]
		return batch, nil
	}
	if ib.done {
		return nil, nil
	}

	// No batch is queued, run a query to get a couple more queued batches
	rows, err := ib.tx.Queryx(fmt.Sprintf("SELECT data FROM %s WHERE tsid=? AND tstart > ? ORDER BY tstart ASC LIMIT ?", ib.table), ib.tsid, ib.tstart, ib.queueSize)
	if err != nil {
		return nil, err
	}
	bi := SQLBatchIterator{rows, nil}
	defer bi.Close()
	for len(ib.queuedbatches) < ib.queueSize {
		batch, err := bi.NextBatch()
		if err != nil {
			return nil, err
		}
		if batch == nil {
			ib.done = true
			return ib.NextBatch()
		}
		ib.tstart = batch[0].Timestamp
		ib.queuedbatches = append(ib.queuedbatches, batch)
	}
	return ib.NextBatch()
}

func NewIteratedBatcher(tx database.TxWrapper, table string, tsid string, tstart float64, queueSize int) *IteratedBatcher {
	return &IteratedBatcher{
		tx:            tx,
		tstart:        tstart,
		tsid:          tsid,
		queueSize:     queueSize,
		done:          false,
		table:         table,
		queuedbatches: []DatapointArray{},
	}
}

func (ts *TimeseriesDB) writeBatch(tx database.TxWrapper, table, tsid string, curBatch DatapointArray) error {
	if len(curBatch) == 0 {
		return nil // Don't write an empty batch
	}
	b, err := curBatch.ToBytes()
	if err != nil {
		return err
	}
	if ts.DB.Assets().Config.Verbose {
		logrus.WithField("timeseries", tsid).Debugln("Writing Batch: ", curBatch.String())
	}
	_, err = tx.Exec(fmt.Sprintf("INSERT OR REPLACE INTO %s(tsid,tstart,tend,length,data) VALUES (?,?,?,?,?);", table), tsid, curBatch[0].Timestamp, curBatch[len(curBatch)-1].EndTime(), len(curBatch), b)
	return err
}

func (ts *TimeseriesDB) writeBatchStart(tx database.TxWrapper, table, tsid string, curBatch DatapointArray) (DatapointArray, error) {
	for len(curBatch) > ts.MaxBatchSize {
		prevBatch := curBatch[:ts.BatchSize]
		curBatch = curBatch[ts.BatchSize:]
		err := ts.writeBatch(tx, table, tsid, prevBatch)
		if err != nil {
			return curBatch, err
		}
	}
	return curBatch, nil
}

// appendUntil takes the current batch, with the assumption that all data in the DatapointIterator comes after the timestamp of the last element of curBatch,
// and keeps appending until the timestamps overlap with until time. The resulting batch will never go beyond until time.
func (ts *TimeseriesDB) appendUntil(tx database.TxWrapper, table, tsid string, curBatch DatapointArray, dp *Datapoint, modified bool, data DatapointIterator, until float64) (*Datapoint, DatapointArray, bool, error) {
	var err error
	for dp != nil && (dp.EndTime() < until || dp.Duration > 0 && dp.EndTime() <= until) {
		modified = true
		curBatch = append(curBatch, dp)
		curBatch, err = ts.writeBatchStart(tx, table, tsid, curBatch)
		if err != nil {
			return nil, nil, false, err
		}
		dp, err = data.Next()
		if err != nil {
			return nil, nil, false, err
		}
	}

	return dp, curBatch, modified, nil
}

var errConflict = errors.New("bad_query: conflict with existing datapoint")

// mergeBatch merges the data in the given batch with the new data from iterator. It guarantees that once finished,
// the end time of the resulting batch never exceeds the end time of the original batch
func (ts *TimeseriesDB) mergeBatch(tx database.TxWrapper, table, tsid string, cb DatapointArray, modif bool, method int, dpi DatapointIterator, idp *Datapoint) (dp *Datapoint, curBatch DatapointArray, modified bool, err error) {
	dp = idp
	modified = modif
	curBatch = cb
	if dp.Timestamp >= curBatch[len(curBatch)-1].EndTime() {
		if dp.Timestamp == curBatch[len(curBatch)-1].Timestamp {
			if dp.Duration == 0 {
				if !dp.IsEqual(curBatch[len(curBatch)-1]) {
					modified = true
					if method > 0 {
						err = errConflict
						return
					}
					curBatch[len(curBatch)-1] = dp
				}
				dp, err = dpi.Next()
			} else {
				modified = true
				curBatch = curBatch[:len(curBatch)-1]
			}

		}
		return
	}

	cbi := NewDatapointArrayIterator(curBatch)
	cdp, _ := cbi.Next()

	curBatch = make(DatapointArray, 0, ts.MaxBatchSize+1)

	for {
		// Append to the batch until cdp reaches dp
		for dp == nil || cdp.Timestamp < dp.Timestamp {
			curBatch = append(curBatch, cdp)

			cdp, _ = cbi.Next()
			if cdp == nil {
				// We're done! Need to check for conflict on the last datapoint though
				if dp != nil && curBatch[len(curBatch)-1].EndTime() > dp.Timestamp {
					// Yeah, remove the last datapoint
					curBatch = curBatch[:len(curBatch)-1]
					modified = true

					if method > 0 {
						err = errConflict
					}
				}
				return
			}

		}
		// check if there is a conflict with end datapoint, and remove it if method is insert
		if len(curBatch) > 0 && curBatch[len(curBatch)-1].EndTime() > dp.Timestamp {
			// Yeah, remove the last datapoint
			curBatch = curBatch[:len(curBatch)-1]
			modified = true

			if method > 0 {
				err = errConflict
				return
			}
		}

		// Add all points that fit into current point's location
		for dp != nil && dp.Timestamp <= cdp.Timestamp && dp.EndTime() <= cdp.EndTime() {
			if !dp.IsEqual(cdp) {
				modified = true
				if method > 0 {
					err = errConflict
					return
				}
			}

			curBatch = append(curBatch, dp)
			curBatch, err = ts.writeBatchStart(tx, table, tsid, curBatch)
			if err != nil {
				return
			}
			for dp.Timestamp == cdp.Timestamp || dp.EndTime() > cdp.Timestamp {
				if method > 0 {
					err = errConflict
					return
				}
				modified = true
				cdp, _ = cbi.Next()
				if cdp == nil {
					dp, err = dpi.Next()
					return
				}
			}

			dp, err = dpi.Next()
			if err != nil {
				return
			}
		}
		for dp != nil && dp.Timestamp <= cdp.Timestamp && dp.EndTime() > cdp.EndTime() {
			// If the timestamp goes *beyond* cdp, it means that we skip cdp
			if method > 0 {
				err = errConflict
				return
			}
			modified = true
			cdp, _ = cbi.Next()
			if cdp == nil {
				return
			}
		}
	}

}

type batchinfo struct {
	tstart float64
	tend   float64
	length int
	data   []byte
}

func (ts *TimeseriesDB) append(tx database.TxWrapper, table, tsid string, curBatch DatapointArray, data DatapointIterator, dp *Datapoint) error {
	// This is an appending insert. Let's DO THIS, we are now free to go crazy - we can prepare the batches in another thread entirely,
	// and just use this thread for pure database writes. This helps because in general json marshalling and gzipping takes some time

	var gerr error
	closer := make(chan bool, 1)
	batcher := make(chan *batchinfo, 3)

	go func() {
		for {
			if dp == nil {
				b, err := curBatch.ToBytes()
				if err != nil {
					gerr = err
					batcher <- nil
					return
				}
				if ts.DB.Assets().Config.Verbose {
					logrus.WithField("timeseries", tsid).Debugln("Appending Batch: ", curBatch.String())
				}
				// Write the remaining elements of this batch, and exit
				batcher <- &batchinfo{
					tstart: curBatch[0].Timestamp,
					tend:   curBatch[len(curBatch)-1].Timestamp + curBatch[len(curBatch)-1].Duration,
					length: len(curBatch),
					data:   b,
				}
				batcher <- nil
				return
			}

			curBatch = append(curBatch, dp)

			if len(curBatch) > ts.MaxBatchSize {
				prevBatch := curBatch[:ts.BatchSize]
				curBatch = curBatch[ts.BatchSize:]
				b, err := prevBatch.ToBytes()
				if err != nil {
					gerr = err
					batcher <- nil
					return
				}
				if ts.DB.Assets().Config.Verbose {
					logrus.WithField("timeseries", tsid).Debugln("Appending Batch: ", prevBatch.String())
				}
				select {
				case <-closer:
					batcher <- nil
					return
				case batcher <- &batchinfo{
					tstart: prevBatch[0].Timestamp,
					tend:   prevBatch[len(prevBatch)-1].Timestamp + prevBatch[len(prevBatch)-1].Duration,
					length: len(prevBatch),
					data:   b,
				}:

				}

			}

			dp, gerr = data.Next()
			if gerr != nil {
				batcher <- nil
				return
			}
		}

	}()

	statement := fmt.Sprintf("INSERT OR REPLACE INTO %s(tsid,tstart,tend,length,data) VALUES (?,?,?,?,?);", table)

	for b := <-batcher; b != nil; b = <-batcher {
		_, err := tx.Exec(statement, tsid, b.tstart, b.tend, b.length, b.data)
		if err != nil {
			closer <- true
			for b = <-batcher; b != nil; b = <-batcher {
			} // Wait until batcher closes
			return err
		}
	}
	return gerr
}

type InsertQuery struct {
	Actions  *bool `json:"actions,omitempty"`
	Validate *bool `json:"validate,omitempty"` // Whether or not to validate the insert against the schema

	// insert, append, update - default is update
	Method *string `json:"method,omitempty"`
}

func (ts *TimeseriesDB) Insert(tsid string, data DatapointIterator, q *InsertQuery) (err error) {
	table := "timeseries"
	method := 0 // 0 is update

	// Make sure data comes in sorted and without any funny business
	data = NewSortChecker(data)

	if q != nil {
		if q.Actions != nil && *q.Actions {
			table = "timeseries_actions"
		}

		if q.Method != nil {
			if *q.Method == "insert" {
				method = 1
			} else if *q.Method == "append" {
				method = 2
			} else if *q.Method == "update" {
			} else {
				return errors.New("bad_query: Unrecognized insert method")
			}
		}
	}

	delStatement := fmt.Sprintf("DELETE FROM %s WHERE tsid=? AND tstart=?", table)

	var dp *Datapoint
	dp, err = data.Next()
	if err != nil || dp == nil {
		return err
	}

	var tx database.TxWrapper
	tx, err = ts.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Get the batch immediately preceding the datapoint
	var rows *sqlx.Rows
	rows, err = tx.Queryx(fmt.Sprintf("SELECT data FROM %s WHERE tsid=? AND tstart <= ? ORDER BY tstart DESC LIMIT 1", table), tsid, dp.Timestamp)
	if err != nil {
		return err
	}
	bi := SQLBatchIterator{rows, nil}

	curBatch, err := bi.NextBatch()
	if err != nil {
		bi.Close()
		return err
	}
	bi.Close()
	modified := false

	// The iteratedBatcher won't actually query the database until we run NextBatch()
	ib := NewIteratedBatcher(tx, table, tsid, dp.Timestamp, 5)
	defer ib.Close()

	if curBatch == nil {
		// If just starting out, initialize curBatch as an empty array
		curBatch = DatapointArray{}
	} else {
		// fmt.Println("cb", curBatch.String(), "dp", dp)
		// curBatch is not nil, merge the data in the batch
		cbstart := curBatch[0].Timestamp
		dp, curBatch, modified, err = ts.mergeBatch(tx, table, tsid, curBatch, modified, method, data, dp)
		if err != nil {
			return err
		}
		if modified && method >= 2 {
			return errConflict
		}
		if dp == nil { // Merging was enough to finish the data
			if modified {
				return ts.writeBatch(tx, table, tsid, curBatch)
			}
			return nil
		}
		if len(curBatch) == 0 {
			// The batch is empty... That means that the entire batch was replaced, so make sure to delete it
			_, err = tx.Exec(delStatement, tsid, cbstart)
			if err != nil {
				return err
			}
		}
	}
	for {
		// fmt.Println("cb", curBatch.String(), "dp", dp)
		nextBatch, err := ib.NextBatch()
		if err != nil {
			return err
		}
		// fmt.Println("nb", nextBatch.String())
		if nextBatch == nil {
			break // We are at the end of the series. any further inserts are pure appends.
		}

		// Then append data between batches
		dp, curBatch, modified, err = ts.appendUntil(tx, table, tsid, curBatch, dp, modified, data, nextBatch[0].Timestamp)
		if err != nil {
			return err
		}
		if modified && method > 0 {
			return errConflict
		}
		if dp == nil { // No need for more data, we're finished
			if modified {
				return ts.writeBatch(tx, table, tsid, curBatch)
			}
			return nil
		}

		// Delete all batches that are entirely within the range of next datapoint
		for nextBatch[0].Timestamp > dp.Timestamp && nextBatch[len(nextBatch)-1].Timestamp < dp.EndTime() {
			if method > 0 {
				return errConflict
			}

			_, err = tx.Exec(delStatement, tsid, nextBatch[0].Timestamp)
			if err != nil {
				return err
			}
			nextBatch, err = ib.NextBatch()
			if err != nil {
				return err
			}
			if nextBatch == nil {
				break // We are at the end of the series.
			}
			// Then append data between batches again
			dp, curBatch, modified, err = ts.appendUntil(tx, table, tsid, curBatch, dp, modified, data, nextBatch[0].Timestamp)
			if err != nil {
				return err
			}
			if dp == nil { // No need for more data, we're finished
				if modified {
					return ts.writeBatch(tx, table, tsid, curBatch)
				}
				return nil
			}
		}
		if nextBatch == nil {
			break
		}

		// If the next datapoint starts before next batch, but it overlaps it, merge the batches and redo procedure
		if nextBatch[0].Timestamp > dp.Timestamp {
			// Here, we want to delete the future batch, and append it to current batch - then rerunning the merge code will handle
			// rebatching based on new data, since there is an overlap in the batches
			if method > 0 {
				return errConflict
			}

			_, err = tx.Exec(delStatement, tsid, nextBatch[0].Timestamp)
			if err != nil {
				return err
			}

			i := 0
			if len(curBatch) > 0 {
				for ; i < len(nextBatch) && (nextBatch[i].Timestamp < curBatch[len(curBatch)-1].EndTime() || curBatch[len(curBatch)-1].Duration == 0 && nextBatch[i].Timestamp == curBatch[len(curBatch)-1].Timestamp); i++ {
				}
			}

			// modified = true - if we're here, we already know stuff was modified
			curBatch = append(curBatch, nextBatch[i:]...)
		} else {
			// The next datapoint will be handled entirely by the next batch
			if modified {
				err = ts.writeBatch(tx, table, tsid, curBatch)
				if err != nil {
					return err
				}
			}
			modified = false
			curBatch = nextBatch
		}

		// curBatch is not nil, merge the data in the batch (here curBatch is guaranteed to be not nil)
		cbstart := curBatch[0].Timestamp
		dp, curBatch, modified, err = ts.mergeBatch(tx, table, tsid, curBatch, modified, method, data, dp)
		if err != nil {
			return err
		}
		if modified && method >= 2 {
			return errConflict
		}
		if dp == nil { // Merging was enough to finish the data
			if modified {
				return ts.writeBatch(tx, table, tsid, curBatch)
			}
			return nil
		}
		if len(curBatch) == 0 {
			// The batch is empty... That means that the entire batch was replaced, so make sure to delete it
			_, err = tx.Exec(delStatement, tsid, cbstart)
			if err != nil {
				return err
			}
		}
	}

	// Whew, we are now at the end of the existing timeseries - this means we are appending, so no more need to worry about merging with existing data.
	return ts.append(tx, table, tsid, curBatch, data, dp)
}
