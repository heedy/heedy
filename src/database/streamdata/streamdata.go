package streamdata

type Datapoint struct {
	T float64
	D interface{}
}

type DatapointIterator interface {
	Next() (Datapoint, error)
	Close() error
}

type StreamData interface {
	Append(streamID uint64, data DatapointIterator) error
	Insert(StreamID uint64, data DatapointIterator) error
	Remove(StreamID uint64, timestamp float64) error
	RemoveTRange(StreamID uint64, t1 float64, t2 float64) error
	RemoveIRange(StreamID uint64, i1 uint64, i2 uint64) error
	QueryTRange(StreamID uint64, t1 float64, t2 float64, limit int64) (DatapointIterator, error)
	QueryIRange(StreamID uint64, i1 uint64, i2 uint64, limit int64) (DatapointIterator, error)
	Delete(StreamID uint64)
	Create(StreamID uint64)
}
