package datastream

import "database/sql"

//SqlRange is a range object that conforms to the range interface
type SqlRange struct {
	r     *sql.Rows
	da    *DatapointArray
	index int64
}

//Close clears all resources used by the sqlRange
func (s *SqlRange) Close() {
	if s.r != nil {
		s.r.Close()
		s.r = nil
	}
}

//Index returns the current index of the values in the sql range
func (s *SqlRange) Index() int64 {
	return s.index
}

//NextArray returns the next DatapointArray chunk from the database
func (s *SqlRange) NextArray() (da *DatapointArray, err error) {
	//Is there is a current array, return that
	if s.da != nil && s.da.Length() > 0 {
		tmp := s.da
		s.index += int64(s.da.Length())
		s.da = nil
		return tmp, nil
	}

	//Check if the iterator is functional
	if s.r == nil {
		return nil, nil
	}

	if !s.r.Next() { //Check if there is more data to read
		err = s.r.Err()
		s.Close()
		return nil, err
	}

	//There is more data to read!
	var version int
	var endindex int64
	var data []byte
	if err = s.r.Scan(&version, &endindex, &data); err != nil {
		s.Close()
		return nil, err
	}
	if s.da, err = DecodeDatapointArray(data, version); err != nil {
		s.Close()
		return nil, err
	}

	//Repeat the procedure
	return s.NextArray()
}

//Next returns the next datapoint from the range
func (s *SqlRange) Next() (d *Datapoint, err error) {
	if s.da != nil && s.da.Length() > 0 {
		tmp := (*s.da)[0]
		s.da = s.da.IRange(1, s.da.Length())
		s.index++
		return &tmp, nil
	}

	s.da, err = s.NextArray()
	if err != nil {
		return nil, err
	}
	if s.da == nil {
		return nil, nil
	}

	//We need to correct the index because we didn't actually give the array out
	s.index -= int64(s.da.Length())

	return s.Next()
}
