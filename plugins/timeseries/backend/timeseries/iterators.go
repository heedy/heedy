package timeseries

import (
	"database/sql"
	"errors"
	"math"

	"github.com/heedy/pipescript"
	"github.com/jmoiron/sqlx"
	"github.com/xeipuuv/gojsonschema"
)

// BatchIterator iterates throguh successive batches of datapoints coming from the database
type BatchIterator interface {
	NextBatch() (DatapointArray, error)
	Close() error
}

// SQLBatchIterator takes a query that returns the raw batch bytes, and outputs the resulting DatapointArray
type SQLBatchIterator struct {
	Rows   *sqlx.Rows
	Closer func()
}

func (b SQLBatchIterator) Close() error {
	if b.Closer != nil {
		defer b.Closer()
	}
	return b.Rows.Close()
}

func (b SQLBatchIterator) NextBatch() (DatapointArray, error) {
	if !b.Rows.Next() {
		b.Rows.Close()
		return nil, nil
	}
	var raw sql.RawBytes
	err := b.Rows.Scan(&raw)
	if err != nil {
		b.Rows.Close()
		return nil, err
	}
	return DatapointArrayFromBytes(raw)
}

// ChanBatch runs NextBatch() in a goroutine, so that post-processing and pre-processing can happen in parallel
type ChanBatchIterator struct {
	closer      chan bool
	datapointer chan DatapointArray
	err         error
}

func (c *ChanBatchIterator) Close() error {
	if c.closer != nil {
		c.closer <- true
		c.closer = nil
	}
	return nil
}
func (c *ChanBatchIterator) NextBatch() (DatapointArray, error) {
	dp := <-c.datapointer
	if dp == nil {
		return dp, c.err
	}
	return dp, nil
}

func NewChanBatchIterator(di BatchIterator) *ChanBatchIterator {
	closer := make(chan bool)
	datapointer := make(chan DatapointArray, 5)
	ci := &ChanBatchIterator{
		closer:      make(chan bool, 1),
		datapointer: datapointer,
		err:         nil,
	}

	go func() {
		defer di.Close()
		for {
			dp, err := di.NextBatch()
			if err != nil {
				ci.err = err
				dp = nil
			}

			select {
			case datapointer <- dp:
			case <-closer:
				close(datapointer)
				return
			}
			if dp == nil {
				return
			}
		}
	}()
	return ci
}

// BatchEndTime only returns batches with timestamps < endtime
type BatchEndTime struct {
	BatchIterator
	EndTime float64
}

func (bet BatchEndTime) NextBatch() (DatapointArray, error) {
	nb, err := bet.BatchIterator.NextBatch()
	if err != nil || nb == nil || len(nb) == 0 {
		return nb, err
	}
	if nb[len(nb)-1].Timestamp >= bet.EndTime {
		// The batch ends after the end time! Remove all datapoints that are beyond the end time
		i := len(nb) - 1
		for ; i >= 0 && nb[i].Timestamp >= bet.EndTime; i-- {
		}
		if i < 0 {
			return nil, nil
		}
		nb = nb[:i+1]
	}
	return nb, nil
}

type BatchEndOffset struct {
	BatchIterator
	EndBatch float64
	Offset   int
}

func (beo BatchEndOffset) NextBatch() (DatapointArray, error) {
	nb, err := beo.BatchIterator.NextBatch()
	if err != nil || nb == nil || len(nb) == 0 {
		return nb, err
	}
	if nb[0].Timestamp < beo.EndBatch {
		return nb, nil
	}
	if nb[0].Timestamp == beo.EndBatch {
		return nb[:beo.Offset], nil
	}
	return nil, nil
}

type BatchPointLimit struct {
	BatchIterator
	Limit int64
}

func (bpl *BatchPointLimit) NextBatch() (DatapointArray, error) {
	nb, err := bpl.BatchIterator.NextBatch()
	if err != nil || nb == nil || len(nb) == 0 {
		return nil, err
	}
	if bpl.Limit < int64(len(nb)) {
		// This batch needs to be truncated
		nb = nb[:bpl.Limit]
		if len(nb) == 0 {
			nb = nil
		}
	}
	bpl.Limit -= int64(len(nb))
	return nb, nil
}

func Toffset(dpa DatapointArray, t float64) DatapointArray {
	i := 0
	for ; dpa[i].Timestamp < t && i < len(dpa); i++ {
	}
	return dpa[i:]
}

// BatchTOffset reads the BatchIterator until the given start time, and returns the remainder of
// the batch where the time started.
func BatchTOffset(bi BatchIterator, da DatapointArray, t float64) (DatapointArray, error) {
	var err error
	for len(da) == 0 || da[len(da)-1].Timestamp < t {
		da, err = bi.NextBatch()
		if da == nil || err != nil {
			return da, err
		}
	}
	i := 0
	for ; da[i].Timestamp < t && i < len(da); i++ {
	}
	return da[i:], err
}

type DatapointIterator interface {
	Next() (*Datapoint, error)
	Close() error
}

// DataValidator performs all of the validation necessary on a timeseries for it to conform to a permissions-based
// system. This includes validating the schema and ensuring that the actor is set correctly.
type DataValidator struct {
	data   DatapointIterator
	schema *gojsonschema.Schema
	actor  string
}

// NewDataValidator ensures that the timeseries data fits the given schema and has actor set properly
func NewDataValidator(data DatapointIterator, schema interface{}, actor string) (*DataValidator, error) {
	s, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
	if err != nil {
		return nil, err
	}
	return &DataValidator{
		data:   data,
		schema: s,
		actor:  actor,
	}, nil
}

// Next sets the actor, and performs schema validation
func (s *DataValidator) Next() (*Datapoint, error) {
	dp, err := s.data.Next()
	if dp == nil || err != nil {
		return dp, err
	}

	result, err := s.schema.Validate(gojsonschema.NewGoLoader(dp.Data))
	if err != nil {
		s.data.Close()
		return dp, err
	}
	if !result.Valid() {
		s.data.Close()
		return dp, errors.New("bad_query: The data failed schema validation")
	}
	dp.Actor = s.actor
	return dp, nil
}

// Close closes the underlying timeseries
func (s *DataValidator) Close() error {
	return s.data.Close()
}

type InfoIterator struct {
	DatapointIterator
	Tstart    float64
	Tend      float64
	Count     int64
	LastPoint *Datapoint
}

func (i *InfoIterator) Next() (*Datapoint, error) {
	dp, err := i.DatapointIterator.Next()
	if err != nil || dp == nil {
		return dp, err
	}
	i.Count++
	i.Tend = dp.EndTime()
	if math.IsInf(i.Tstart, -1) {
		i.Tstart = dp.Timestamp
	}

	return dp, nil

}

func NewInfoIterator(di DatapointIterator) *InfoIterator {
	return &InfoIterator{
		DatapointIterator: di,
		Tstart:            math.Inf(-1),
		Tend:              math.Inf(-1),
		Count:             0,
		LastPoint:         nil,
	}
}

type SortChecker struct {
	DatapointIterator
	endTime   float64
	inclusive bool
}

func (o *SortChecker) Next() (dp *Datapoint, err error) {
	dp, err = o.DatapointIterator.Next()
	if err != nil || dp == nil {
		return
	}
	if dp.Duration < 0 {
		err = errors.New("bad_query: durations can't be negative")
		return
	}
	if o.inclusive && dp.Timestamp < o.endTime || !o.inclusive && dp.Timestamp <= o.endTime {
		err = errors.New("bad_query: data must be ordered with increasing timestamps, and durations must not intersect")
		return
	}
	o.inclusive = dp.Duration > 0
	o.endTime = dp.EndTime()
	return
}

func NewSortChecker(di DatapointIterator) *SortChecker {
	return &SortChecker{
		DatapointIterator: di,
		endTime:           math.Inf(-1),
		inclusive:         false,
	}
}

// ChanIterator runs the iteration in a goroutine, so that post-processing data and pre-processing
// can happen in parallel
type ChanIterator struct {
	closer      chan bool
	datapointer chan *Datapoint
	err         error
}

func (c *ChanIterator) Close() error {
	if c.closer != nil {
		c.closer <- true
		c.closer = nil
	}
	return nil
}

func (c *ChanIterator) Next() (*Datapoint, error) {
	dp := <-c.datapointer
	if dp == nil {
		return dp, c.err
	}
	return dp, nil
}

func NewChanIterator(di DatapointIterator) *ChanIterator {
	closer := make(chan bool)
	datapointer := make(chan *Datapoint, 10000)
	ci := &ChanIterator{
		closer:      make(chan bool, 1),
		datapointer: datapointer,
		err:         nil,
	}

	go func() {
		defer di.Close()
		for {
			dp, err := di.Next()
			if err != nil {
				ci.err = err
				dp = nil
			}

			select {
			case datapointer <- dp:
			case <-closer:
				close(datapointer)
				return
			}
			if dp == nil {
				return
			}
		}
	}()
	return ci
}

type BatchDatapointIterator struct {
	BatchIterator
	da DatapointArray
	i  int
}

func (bdi *BatchDatapointIterator) Next() (dp *Datapoint, err error) {
	if bdi.i < len(bdi.da) {
		dp = bdi.da[bdi.i]
		bdi.i++
		return dp, nil
	}
	bdi.i = 1

	bdi.da, err = bdi.NextBatch()
	if err != nil || bdi.da == nil {
		return nil, err
	}

	return bdi.da[0], nil
}

func NewBatchDatapointIterator(bi BatchIterator, da DatapointArray) *BatchDatapointIterator {
	return &BatchDatapointIterator{
		BatchIterator: bi,
		da:            da,
		i:             0,
	}
}

type EmptyIterator struct{}

func (e EmptyIterator) Close() error {
	return nil
}

func (e EmptyIterator) Next() (*Datapoint, error) {
	return nil, nil
}

// ------------------------------------------------------------------------------------------------------

//DatapointArrayIterator allows DatapointArray to conform to the DatapointIterator interface
type DatapointArrayIterator struct {
	rangeindex int
	da         DatapointArray
}

//Close resets the range
func (d *DatapointArrayIterator) Close() error {
	d.rangeindex = 0
	return nil
}

//Index returns the index of the DatapointArray
func (d *DatapointArrayIterator) Index() int64 {
	return int64(d.rangeindex)
}

//Next returns the next datapoint
func (d *DatapointArrayIterator) Next() (*Datapoint, error) {
	if d.rangeindex >= len(d.da) {
		return nil, nil
	}
	d.rangeindex++
	return d.da[d.rangeindex-1], nil
}

//NextArray returns what is left of the array
func (d *DatapointArrayIterator) NextArray() (DatapointArray, error) {
	if d.rangeindex >= len(d.da) {
		return nil, nil
	}
	dpa := d.da[d.rangeindex:]
	d.rangeindex = len(d.da)
	return dpa, nil
}

//NewDatapointArrayIterator does exactly what the function says
func NewDatapointArrayIterator(da DatapointArray) *DatapointArrayIterator {
	return &DatapointArrayIterator{0, da}
}

// ------------------------------------------------------------------------------------------------------

//NumIterator returns only the first given number of datapoints (with an optional skip param) from a DatapointIterator
type NumIterator struct {
	di      DatapointIterator
	numleft int64 //The number of datapoints left to return
}

//Close closes the internal DatapointIterator
func (r *NumIterator) Close() error {
	return r.di.Close()
}

//Next returns the next datapoint from the underlying DatapointIterator so long as the datapoint is within the
//amonut of datapoints to return.
func (r *NumIterator) Next() (*Datapoint, error) {
	if r.numleft == 0 {
		r.di.Close()
		return nil, nil
	}
	r.numleft--
	return r.di.Next()
}

//Skip the given number of datapoints without changing the number of datapoints left to return
func (r *NumIterator) Skip(num int) error {
	for i := 0; i < num; i++ {
		_, err := r.di.Next()
		if err != nil {
			return err
		}
	}
	return nil
}

//NewNumIterator initializes a new NumIterator which will return up to the given amount of datapoints.
func NewNumIterator(dr DatapointIterator, datapoints int64) *NumIterator {
	return &NumIterator{dr, datapoints}
}

// ------------------------------------------------------------------------------------------------------

// NewArrayFromIterator creates a datapoint array from the given iterator
func NewArrayFromIterator(di DatapointIterator) (DatapointArray, error) {
	d := DatapointArray{}

	dp, err := di.Next()
	for dp != nil && err == nil {
		d = append(d, dp)
		dp, err = di.Next()
	}
	return d, err
}

// PipeIterator allows using PipeScript transforms over DatapointIterator
type PipeIterator struct {
	it DatapointIterator
}

func (pi PipeIterator) Next(out *pipescript.Datapoint) (*pipescript.Datapoint, error) {
	dp, err := pi.it.Next()
	if dp == nil || err != nil {
		return nil, err
	}
	out.Timestamp = dp.Timestamp
	out.Duration = dp.Duration
	out.Data = dp.Data
	return out, nil
}

type Closer interface {
	Close() error
}

type TransformIterator struct {
	dpi Closer
	it  pipescript.Iterator
	dp  pipescript.Datapoint
}

func (pi *TransformIterator) Next() (*Datapoint, error) {
	dp, err := pi.it.Next(&pi.dp)
	if dp == nil || err != nil {
		return nil, err
	}
	return &Datapoint{
		Timestamp: dp.Timestamp,
		Duration:  dp.Duration,
		Data:      dp.Data,
	}, nil
}

func (pi *TransformIterator) Close() error {
	return pi.dpi.Close()
}

func NewTransformIterator(transform string, it DatapointIterator) (DatapointIterator, error) {
	p, err := pipescript.Parse(transform)
	if err != nil {
		return nil, err
	}
	p.InputIterator(PipeIterator{it})
	return &TransformIterator{dpi: it, it: p}, nil
}
