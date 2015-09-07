package transforms

import "connectordb/streamdb/datastream"

func CopyDatapoint(dp *datastream.Datapoint) *datastream.Datapoint {
	return dp.Copy()
}

//DatapointTransform is an interface that transforms one Datapoint at a time. It is guaranteed
//to be called ordered by Datapoints in the stream, so state is allowed to be kept.
//To allow more complicated states, once the ExtendedDataRange runs out of data, a nil is passed through
//the transform until the transform returns nil, to allow internally queued Datapoints to be returned.
//To filter datapoins, returning a null Datapoint without error means the daatpoint was filtered (or internally cached)
type DatapointTransform interface {
	Transform(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error)
}

// A straightforward wrapper for functions that adhere to DatapointTransform
type DatapointTransformWrapper struct {
	Transformer TransformFunc
}

func (d DatapointTransformWrapper) Transform(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
	input := NewTransformEnvironment(dp)
	out := d.Transformer(input)
	return out.Datapoint, out.Error
}

// Creates a new transform pipeline from the given pipeline definition
func NewTransformPipeline(pipeline string) (DatapointTransform, error) {
	transformer, err := ParseTransform(pipeline)

	return DatapointTransformWrapper{transformer}, err
}

//go:generate go tool yacc -o transform_generator_y.go -p Transform pipeline_generator.y
//
// type Transformer struct {
// 	err      error
// 	pipeline TransformFunc
// 	input    <-chan *datastream.Datapoint
// }
//
// func (t *Transformer) Next() <-chan *datastream.Datapoint {
// 	ch := make(chan *datastream.Datapoint, 10)
// 	go func() {
// 		// initialize vars.
// 		current := <-t.input
// 		next := <-t.input
//
// 		for current != nil {
// 			val := NewTransformEnvironment(current)
//
// 			// If this is the last non-nil datapoint we see, set the flag
// 			// so the system knows to clean up.
// 			if next == nil {
// 				val.Flag = LastDatapoint
// 			}
//
// 			t.pipeline(val)
//
// 			// on error, we close the pipe
// 			t.err = val.Error
// 			if t.err != nil {
// 				return
// 			}
//
// 			if val.Datapoint != nil {
// 				ch <- val.Datapoint
// 			}
//
// 			current = next
// 			next = <-t.input
// 		}
//
// 		close(ch)
// 	}()
//
// 	return ch
// }
//
// // Returns the error associated with this transformer
// func (t *Transformer) Error() error {
// 	return t.err
// }
//
// func newTransformer(pipeline string, input <-chan *datastream.Datapoint) (*Transformer, error) {
// 	transformer, err := ParseTransform(pipeline)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &Transformer{nil, transformer, input}, nil
// }
//
// // Creates a new transform pipeline from the given pipeline definition
// func NewTransformPipeline(pipeline string, values []datastream.Datapoint) (*Transformer, error) {
//
// 	ch := make(chan *datastream.Datapoint)
// 	go func() {
// 		for _, val := range values {
// 			ch <- &val
// 		}
// 		close(ch)
// 	}()
//
// 	return newTransformer(pipeline, ch)
// }
//
// // Creates a new transform pipeline from the given pipeline definition and a channel of inputs
// func NewChanTransformPipeline(pipeline string, values <-chan *datastream.Datapoint) (*Transformer, error) {
// 	return newTransformer(pipeline, values)
// }
