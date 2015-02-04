package timebatchdb

import (
    "container/list"
    )

//The DataRange interface - this is the object that is returned from different caches/stores - it represents
//a range of data values stored in a certain way, and Next() gets the next datapoint in the range
type DataRange interface {
    Init()              //Does the necessary steps to get the datarange ready for returning datapoints
    Next() *Datapoint   //Returns the next datapoint in sequence - or nil if the sequence is finished
    Close()             //Closes the datarange - can be called before Init. But Init does not have to work after close.
}


//The RangeList - it is a list of DataRanges, and acts as one large DataRange. In particular, it can combine
//datasets with certain data overlap, since it makes sure that timestamps are strictly increasing.
type RangeList struct {
    rlist *list.List    //A list of DataRange objects
    prevpt *Datapoint   //The previous datapoint - to make sure timestamps are ordered
}

func (r *RangeList) Init() {
    if (r.rlist.Len()!=0) {
        //Initialize the first in the list
        r.rlist.Front().Value.(DataRange).Init()
    }
}

func (r *RangeList) Close() {
    if (r.rlist.Len()>0) {
        //Closes all child DataRanges
        elem := r.rlist.Front()
        for elem.Next()!=nil {
            elem.Value.(DataRange).Close()
            elem = elem.Next()
        }
        elem.Value.(DataRange).Close()
    }
}

//Returns the next available value from the list, initializing and closing the necessary stuff
func (r *RangeList) getnextvalue() *Datapoint {
    if (r.rlist.Len()==0) {
        return nil
    }
    e := r.rlist.Front().Value.(DataRange)
    d := e.Next()
    if (d!=nil) {
        return d
    }

    //Okay, this element of the list is empty, we close it, remove it from the list,
    //initialize the next element, and repeat
    e.Close()
    r.rlist.Remove(r.rlist.Front())
    if (r.rlist.Len()==0) {
        return nil
    }
    //Initialize the next element
    r.rlist.Front().Value.(DataRange).Init()

    //repeat the procedure
    return r.getnextvalue()

}

//Returns the next datapoint
func (r *RangeList) Next() *Datapoint {
    d := r.getnextvalue()
    if d==nil {
        return d
    }
    if (r.prevpt==nil) {
        r.prevpt = d
        return d
    }

    //A previous datapoint exists: make sure that the next datapoint
    //has a timestamp > than the last timestamp
    //BUG(daniel): There is no way to check multiple points with same time stamp
    for d.Timestamp() <= r.prevpt.Timestamp() {
        d = r.getnextvalue()
        if (d==nil) {
            return d
        }
    }
    r.prevpt = d
    return d
}

//Appends to the end of the rangelist an uninitialized datarange
func (r *RangeList) Append(d DataRange) {
    r.rlist.PushBack(d)
}

//Creates empty RangeList
func NewRangeList() *RangeList {
    return &RangeList{list.New(),nil}
}
