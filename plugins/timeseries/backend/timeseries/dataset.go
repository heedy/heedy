package timeseries

import (
	"errors"
	"fmt"
	"math"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/pipescript"
	"github.com/heedy/pipescript/datasets"
)

type DatasetIterator struct {
	closers []Closer
	it      pipescript.Iterator
}

func (di *DatasetIterator) Next(out *pipescript.Datapoint) (*pipescript.Datapoint, error) {
	return di.it.Next(out)
}

func (di *DatasetIterator) Close() (err error) {
	for _, c := range di.closers {
		err2 := c.Close()
		if err2 != nil {
			err = err2
		}
	}
	return err
}

func (q *Query) Get(db database.DB, tstart float64) (*DatasetIterator, error) {
	if q.T1 == nil && q.I1 == nil && q.T == nil && q.I == nil {
		q.T1 = tstart
	}

	obj, err := db.ReadObject(q.Timeseries, &database.ReadObjectOptions{
		Icon: false,
	})
	if err != nil {
		return nil, err
	}
	if *obj.Type != "timeseries" {
		return nil, fmt.Errorf("bad_query: Object '%s' is not a timeseries", q.Timeseries)
	}
	if !obj.Access.HasScope("read") {
		return nil, errors.New("access_denied: The given object can't be read")
	}

	iter, err := TSDB.Query(q)
	if err != nil {
		return nil, err
	}

	return &DatasetIterator{
		closers: []Closer{iter},
		it:      PipeIterator{iter},
	}, nil
}

func GetMerge(db database.DB, q []*Query, tstart float64) (*DatasetIterator, error) {
	closers := make([]Closer, len(q))
	iters := make([]pipescript.Iterator, len(q))
	for i := range q {
		it, err := q[i].Get(db, tstart)
		if err != nil {
			for j := 0; j < i; j++ {
				closers[j].Close()
			}
			return nil, err
		}
		closers[i] = it
		iters[i] = it
	}
	it, err := datasets.Merge(iters)
	if err != nil {
		for i := range closers {
			closers[i].Close()
		}
		return nil, err
	}
	return &DatasetIterator{
		closers: closers,
		it:      it,
	}, nil
}

func mergeTimeseries(q []*Query) map[string]int {
	m := make(map[string]int)
	for _, qi := range q {
		cv, ok := m[qi.Timeseries]
		if !ok {
			cv = 0
		}
		m[qi.Timeseries] = cv + 1
	}
	return m
}

type DatasetElement struct {
	Query
	Merge        []*Query `json:"merge"`
	Interpolator string   `json:"interpolator"`
	AllowNull    bool     `json:"allow_null"`
}

func (d *DatasetElement) GetTimeseries() map[string]int {
	if d.Timeseries != "" {
		m := make(map[string]int)
		m[d.Timeseries] = 1
		return m
	}
	return mergeTimeseries(d.Merge)
}

func (d *DatasetElement) Validate() error {
	if d.Query.Timeseries != "" && len(d.Merge) > 0 {
		return errors.New("bad_query: Can't create dataset using both merge and timeseries")
	}
	for _, q := range d.Merge {
		if q.Timeseries == "" {
			return errors.New("bad_query: timseries not specified for merge")
		}
		// Allow specifying a single time range for an entire merge query
		if q.T1 == nil && d.T1 != nil {
			q.T1 = d.T1
		}
		if q.T2 == nil && d.T2 != nil {
			q.T2 = d.T2
		}
		if q.T == nil && q.T != nil {
			q.T = d.T
		}
	}
	if d.Interpolator == "" {
		d.Interpolator = "closest"
	}
	return nil
}

func (d *DatasetElement) Get(db database.DB, tstart float64) (*DatasetIterator, error) {
	if len(d.Merge) > 0 {
		m, err := GetMerge(db, d.Merge, tstart)
		if err != nil || d.Query.Transform == nil {
			return m, err
		}

		// If there is a transform, it is run on the result of the merge
		q, err := pipescript.Parse(*d.Query.Transform)
		if err != nil {
			m.Close()
			return nil, err
		}
		q.InputIterator(m.it)
		m.it = q
		return m, err
	}
	return d.Query.Get(db, tstart)
}

type Dataset struct {
	Query
	Merge []*Query    `json:"merge,omitempty"`
	Dt    interface{} `json:"dt,omitempty"`
	Key   string      `json:"key,omitempty"`
	// A dataset is a map of subelements, which can themselves be merge queries
	Dataset map[string]*DatasetElement `json:"dataset,omitempty"`

	PostTransform string `json:"post_transform,omitempty"`
}

// GetTimeseries returns a map of all the timeseries IDs included in the query
func (d *Dataset) GetTimeseries() (m map[string]int) {
	if len(d.Merge) > 0 {
		m = mergeTimeseries(d.Merge)
	} else {
		m = make(map[string]int)
	}
	if d.Timeseries != "" {
		m[d.Timeseries] = 1
	}
	for _, v := range d.Dataset {
		m2 := v.GetTimeseries()
		for k, v := range m2 {
			cv, ok := m[k]
			if !ok {
				cv = 0
			}
			m[k] = cv + v
		}
	}

	return
}

func (d *Dataset) Validate() error {
	if d.Query.Timeseries != "" && len(d.Merge) > 0 {
		return errors.New("bad_query: Can't create dataset using both merge and timeseries")
	}
	if d.Dt != nil && (d.Query.Timeseries != "" || len(d.Merge) > 0) {
		return errors.New("bad_query: dt and timeseries/merge cannot be used at the same time")
	}
	if d.Dt != nil && (d.T1 == nil) {
		return errors.New("bad_query: t-dataset must have start time")
	}
	if d.Dt != nil && (d.T2 == nil) {
		d.T2 = "now"
	}
	if len(d.Dataset) == 0 && d.Dt != nil {
		return errors.New("bad_query: Can't query dt without a dataset")
	}
	if d.Query.Timeseries != "" && len(d.Dataset) > 0 {
		if d.Key == "" {
			d.Key = "x"
		}
		_, ok := d.Dataset[d.Key]
		if ok {
			return fmt.Errorf("bad_query: Dataset is already using key '%s'", d.Key)
		}
	}

	for _, q := range d.Merge {
		if q.Timeseries == "" {
			return errors.New("bad_query: timseries not specified for merge")
		}

		// Allow specifying a single time range for an entire merge query
		if q.T1 == nil && d.T1 != nil {
			q.T1 = d.T1
		}
		if q.T2 == nil && d.T2 != nil {
			q.T2 = d.T2
		}
		if q.T == nil && q.T != nil {
			q.T = d.T
		}
	}
	for _, v := range d.Dataset {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dataset) populate(db database.DB, dset *datasets.Dataset, tstart float64) (*DatasetIterator, error) {
	closers := make([]Closer, 0)
	for k, v := range d.Dataset {
		di, err := v.Get(db, tstart)
		if err != nil {
			for j := range closers {
				closers[j].Close()
			}
			return nil, err
		}
		ipltr, err := datasets.GetInterpolator(v.Interpolator, nil, dset.Reference(), di)
		if err != nil {
			for j := range closers {
				closers[j].Close()
			}
			return nil, err
		}
		closers = append(closers, di)
		dset.Add(k, ipltr)
	}
	return &DatasetIterator{
		closers: closers,
		it:      dset,
	}, nil
}

func (d *Dataset) Get(db database.DB) (*DatasetIterator, error) {
	err := d.Validate()
	if err != nil {
		return nil, err
	}
	if d.Dt != nil {
		// It is a t-dataset
		dt, err := ParseTimestamp(d.Dt)
		if err != nil {
			return nil, err
		}
		t1, err := ParseTimestamp(d.T1)
		if err != nil {
			return nil, err
		}
		t2, err := ParseTimestamp(d.T2)
		if err != nil {
			return nil, err
		}
		dset := datasets.NewTDataset(t1, t2, dt)
		di, err := d.populate(db, dset, t1)
		if err != nil {
			return nil, err
		}
		if d.PostTransform != "" {
			p, err := pipescript.Parse(d.PostTransform)
			if err != nil {
				di.Close()
				return nil, err
			}
			p.InputIterator(di.it)
			di.it = p
		}
		return di, err
	}
	// Otherwise, it is either just a query, or a dataset. Either way, get the query.
	// We use a trick by just using DatasetElement code here
	di, err := (&DatasetElement{
		Query: d.Query,
		Merge: d.Merge,
	}).Get(db, math.Inf(-1))

	if err != nil || len(d.Dataset) == 0 {
		return di, err // if it was just a query, return the result as is
	}

	dset := datasets.NewDataset(di)

	// Add the current stream to the dataset
	dset.Add(d.Key, pipescript.IteratorFromBI{dset.Reference()})

	// Get tstart by peeking at the first output datapoint
	ref := dset.Reference()
	dp, err := ref.Next()
	if err != nil {
		di.Close()
		return nil, err
	}
	ref.Close() // Close the reference, so that buffer doesn't keep all data in memory
	if dp == nil {
		// There was actually no data... so return an empty dataset
		di.Close()
		return &DatasetIterator{
			closers: []Closer{},
			it:      pipescript.NewDatapointArrayIterator([]pipescript.Datapoint{}),
		}, nil
	}
	di2, err := d.populate(db, dset, dp.Timestamp)
	if err != nil {
		return nil, err
	}
	// Add di to the closers of di2
	di2.closers = append(di2.closers, di)

	if d.PostTransform != "" {
		p, err := pipescript.Parse(d.PostTransform)
		if err != nil {
			di2.Close()
			return nil, err
		}
		p.InputIterator(di2.it)
		di2.it = p
	}

	return di2, nil
}
