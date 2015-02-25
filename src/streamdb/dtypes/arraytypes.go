package dtypes

import (
    "errors"
    "math"
    "encoding/binary"
    "streamdb/timebatchdb"
    )

var (
    ERROR_CORRUPTED_TYPES = errors.New("Data corrupted - types don't match")
    ArrayTypes = map[string]DataType{"x": &BinaryType{},
                        "s": &TextType{},
                        "f": &FloatArrayType{},
                        "i": &IntArrayType{}}
)



type FloatArray struct {
    BaseDatapoint
    D []float64
}
func (td FloatArray) Data() []byte {
    d := make([]byte,8*len(td.D))
    for i := 0; i < len(td.D);i++ {
        binary.LittleEndian.PutUint64(d[i*8:(i+1)*8],math.Float64bits(td.D[i]))
    }
    return d
}
func (td *FloatArray) Load(dp timebatchdb.Datapoint) error {
    td.LoadTime(dp.Timestamp())
    d := dp.Data()
    if len(d)%8 != 0 {
        return ERROR_CORRUPTED_TYPES
    }
    td.D = make([]float64,len(d)/8)
    for i := 0; i < len(td.D);i++ {
        td.D[i] = math.Float64frombits(binary.LittleEndian.Uint64(d[i*8:(i+1)*8]))
    }
    return nil
}
type FloatArrayType struct {
    lenlimit int
}
func (t *FloatArrayType) New() TypedDatapoint {return new(FloatArray)}
func (t *FloatArrayType) Len(i int) DataType {
    return &FloatArrayType{i}
}
func (td *FloatArrayType) IsValid(tda TypedDatapoint) bool {
    t := tda.(*FloatArray)
    if td.lenlimit < 0 {
        return -td.lenlimit >= len(t.D)
    } else if td.lenlimit > 0 {
        return td.lenlimit == len(t.D)
    }
    return true
}


type IntArray struct {
    BaseDatapoint
    D []int64
}
func (td IntArray) Data() []byte {
    d := make([]byte,8*len(td.D))
    for i := 0; i < len(td.D);i++ {
        binary.LittleEndian.PutUint64(d[i*8:(i+1)*8],uint64(td.D[i]))
    }
    return d
}
func (td *IntArray) Load(dp timebatchdb.Datapoint) error {
    td.LoadTime(dp.Timestamp())
    d := dp.Data()
    if len(d)%8 != 0 {
        return ERROR_CORRUPTED_TYPES
    }
    td.D = make([]int64,len(d)/8)
    for i := 0; i < len(td.D);i++ {
        td.D[i] = int64(binary.LittleEndian.Uint64(d[i*8:(i+1)*8]))
    }
    return nil
}
type IntArrayType struct {
    lenlimit int
}
func (t *IntArrayType) New() TypedDatapoint {return new(IntArray)}
func (t *IntArrayType) Len(i int) DataType {
    return &IntArrayType{i}
}
func (td *IntArrayType) IsValid(tda TypedDatapoint) bool {
    t := tda.(*IntArray)
    if td.lenlimit < 0 {
        return -td.lenlimit >= len(t.D)
    } else if td.lenlimit > 0 {
        return td.lenlimit == len(t.D)
    }
    return true
}
