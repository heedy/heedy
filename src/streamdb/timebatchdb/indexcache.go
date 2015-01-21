package timebatchdb

import (
    "os"
    )

type IndexCache struct {
    Ip *IndexPage     //Index page - contains the cached page of the index
    batch map[int](*BatchReader) //Batch Reader map - contains the batches which are cached from this index page
}

//Returns the total size in bytes of data currently cached
func (ic *IndexCache) Size() int64 {
    size := int64(0)
    for _,br := range ic.batch {
        size += int64(br.Size())
    }
    return size
}

func (ic *IndexCache) Clear() {
    ic.batch = make(map[int](*BatchReader))
}

//Given the index on our page of the batch to load, and the data/offset file, load the batch into cache
func (ic *IndexCache) LoadBatch(index int, offsetf *os.File, dataf *os.File) (*BatchReader) {
    dloc,dsize,oloc,osize := ic.Ip.GetLocs(index)

    databuffer := make([]byte,dsize)
    dataf.ReadAt(databuffer,int64(dloc))

    offsetbuffer := make([]byte,osize)
    offsetf.ReadAt(offsetbuffer,int64(oloc))

    br := NewBatchReader(offsetbuffer,databuffer)

    //Add to the cache
    ic.batch[index] = br

    return br
}

//Returns a cached batch, or nil if batch is not cached
func (ic *IndexCache) GetBatch(index int) (*BatchReader,bool) {
    a,b := ic.batch[index]
    return a,b
}

//Returns the timestamp, the datanum and previndex of the index element at the given index on the page.
func (ic *IndexCache) GetIndexValues(index int) (timestamp uint64, datanum uint64, previndex uint64) {
    return ic.Ip.Timestamp[index],ic.Ip.Datanum[index],ic.Ip.Previndex[index]
}


func GetIndexCache(indexfile *os.File, startindex int64, pagesize int) (i *IndexCache,err error){
    ip,err := GetIndexPage(indexfile,startindex,pagesize)
    if (err != nil) {
        return nil,err
    }
    return &IndexCache{ip,make(map[int](*BatchReader))},nil
}
