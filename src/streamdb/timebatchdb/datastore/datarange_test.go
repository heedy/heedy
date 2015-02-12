package datastore

import (
    "testing"
    )

func TestRangeList(t *testing.T) {
    //DataRange can't handle same-timestamp values
    //timestamps := []int64{1000,1500,2000,2000,2000,2500,3000,3000,3000}
    timestamps := []int64{1,2,3,4,5,6,3000,3100,3200}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8")}

    da := CreateDatapointArray(timestamps[:5],data[:5])
    db := CreateDatapointArray(timestamps[2:],data[2:])

    rl := NewRangeList()
    rl.Append(da)
    rl.Append(db)
    rl.Init()

    //Using assertData from datapointarray_test
    /*
    if (!assertData(t,DatapointArrayFromDataRange(rl),"fromdatarange")) {
        return
    }
    */
    da=DatapointArrayFromDataRange(rl)
    if (da.Len()!=9) {
        t.Errorf(" DatapointArray length: %d",da.Len())
        return
    }

    timestamps = da.Timestamps()

    if (timestamps[0]!=1 || timestamps[1]!=2 || timestamps[8]!=3200) {
        t.Errorf("Datarange timestamp fail1: %d %d",timestamps[0],timestamps[8])
        return
    }

}

func TestTimeRange(t *testing.T) {
    timestamps := []int64{1,2,3,4,5,6,6,7,8}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8")}

    da := CreateDatapointArray(timestamps,data)

    tr := NewTimeRange(da,3,6)
    defer tr.Close()
    tr.Init()
    dp := tr.Next()
    if (dp==nil || dp.Timestamp()!=4) {
        t.Errorf("TimeRange start time incorrect")
    }
    dp = tr.Next()
    if (dp==nil || dp.Timestamp()!=5) {
        t.Errorf("TimeRange incorrect")
    }
    dp = tr.Next()
    if (dp==nil || dp.Timestamp()!=6) {
        t.Errorf("TimeRange incorrect")
    }
    dp = tr.Next()
    if (dp==nil || dp.Timestamp()!=6) {
        t.Errorf("TimeRange incorrect")
    }
    dp = tr.Next()
    if (dp!=nil) {
        t.Errorf("TimeRange endtime incorrect")
    }
}

func TestNumRange(t *testing.T) {
    timestamps := []int64{1,2,3,4,5,6,6,7,8}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8")}

    da := CreateDatapointArray(timestamps,data)

    tr := NewNumRange(da,5)
    defer tr.Close()
    tr.Init()
    dp := tr.Next()
    if (dp==nil || dp.Timestamp()!=1) {
        t.Errorf("NumRange start time incorrect")
    }
    tr.Skip(2)
    dp = tr.Next()
    if (dp==nil || dp.Timestamp()!=4) {
        t.Errorf("NumRange start time incorrect")
    }
    dp = tr.Next()
    if (dp==nil || dp.Timestamp()!=5) {
        t.Errorf("NumRange incorrect")
    }
    dp = tr.Next()
    if (dp==nil || dp.Timestamp()!=6) {
        t.Errorf("NumRange incorrect")
    }
    dp = tr.Next()
    if (dp==nil || dp.Timestamp()!=6) {
        t.Errorf("NumRange incorrect")
    }
    dp = tr.Next()
    if (dp!=nil) {
        t.Errorf("NumRange endtime incorrect")
    }
}
