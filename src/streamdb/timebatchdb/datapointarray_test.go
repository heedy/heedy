package timebatchdb

import (
    "testing"
    )

//WARNING: This function is used in tests of datarange also
func assertData(t *testing.T,da *DatapointArray,try string) bool {
    if (da.Len()!=9) {
        t.Errorf("%s: DatapointArray length: %d",try,da.Len())
        return false
    }

    timestamps := da.Timestamps()

    if (timestamps[0]!=1000 || timestamps[1]!=1500 || timestamps[8]!=3000) {
        t.Errorf("%s: DatapointArray timestamp fail1: %d %d",try,timestamps[0],timestamps[8])
        return false
    }

    timestamps,data := da.Get()
    if (len(timestamps)!=9 || len(data)!=9) {
        t.Errorf("%s wrong range returned %d %d",len(timestamps),len(data))
        return false
    }
    if (timestamps[0]!=1000 || timestamps[1]!=1500 || timestamps[8]!=3000) {
        t.Errorf("%s: DatapointArray timestamp fail: %d %d",try,timestamps[0],timestamps[8])
        return false
    }

    if (string(data[0])!="test0" || string(data[1])!="test1" || string(data[8])!="test8" || len(data)!=9) {
        t.Errorf("%s: DatapointArray timestamp fail: %d",try,len(timestamps))
        return false
    }
    return true
}

func TestDatapointArray(t *testing.T) {
    timestamps := []uint64{1000,1500,2000,2000,2000,2500,3000,3000,3000}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8")}

    da := CreateDatapointArray(timestamps,data) //This internally tests fromlist

    if (!assertData(t,da,"creation")) {
        return
    }

    //It looks like the basics are working. Now let's test going to bytes and back
    da.Bytes()

    //da was reloaded when Bytes() was called. Make sure things are fine
    if (!assertData(t,da,"nochangebytes")) {
        return
    }

    //Now test da2
    if (!assertData(t,DatapointArrayFromBytes(da.Bytes()),"frombytes")) {
        return
    }

    if (!assertData(t,DatapointArrayFromCompressedBytes(da.CompressedBytes()),"compressed")) {
        return
    }

    //Now check getting by time
    i := da.FindTimeIndex(1200)
    if (i!= 1) {
        t.Errorf("Error in findtimeindex: %d",i)
    }

    i = da.FindTimeIndex(2000)
    if (i!= 5) {
        t.Errorf("Error in findtimeindex: %d",i)
    }

    i = da.FindTimeIndex(3000)
    if (i!= -1) {
        t.Errorf("Error in findtimeindex: %d",i)
        return
    }

    if (da.DatapointTRange(1200,2000).Len()!=4) {
        t.Errorf("Wrong TRange")
        return
    }

    dp := da.Next()
    if (dp == nil || dp.Timestamp()!=1000) {
        t.Errorf("Iterator wrong")
    }
    dp = da.Next()
    if (dp == nil || dp.Timestamp()!=1500) {
        t.Errorf("Iterator wrong")
    }
    da.Next()
    da.Next()
    da.Next()
    da.Next()
    da.Next()
    da.Next()
    dp = da.Next()
    if (dp == nil || dp.Timestamp()!=3000) {
        t.Errorf("Iterator wrong")
    }
    dp = da.Next()
    if (dp != nil) {
        t.Errorf("Iterator wrong")
    }
    da.Reset()
    dp = da.Next()
    if (dp == nil || dp.Timestamp()!=1000) {
        t.Errorf("Iterator wrong")
    }
    da.Reset()
    //Lastly, make sure loading from DataRange is functional
    if (!assertData(t,DatapointArrayFromDataRange(da),"fromdatarange")) {
        return
    }
}
