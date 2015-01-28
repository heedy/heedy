package timebatchdb

import (
    "os"
    "encoding/binary"
    "path"
    )

type KeyMap struct {
    keyfile *os.File            //The file in which the keymap is stored
    keymap map[string]uint64    //The map itself
    keynum uint64               //The number of keys
}

//Closes the KeyMap
func (k *KeyMap) Close() {
    k.keyfile.Close()
}

//Returns the number of keys in the keymap
func (k *KeyMap) Len() uint64 {
    return k.keynum
}

//Clears the keys, and reloads the entire keymap from file
func (k *KeyMap) Reload() (err error) {
    k.keymap = make(map[string]uint64)

    k.keyfile.Seek(0,0)

    //Read the keys until EOF
    err = nil
    key := uint64(0)
    var keylen uint32
    var keystr []byte
    for err == nil {
        err = binary.Read(k.keyfile,binary.LittleEndian, &keylen)
        if err==nil {
            keystr = make([]byte,keylen)
            _, err := k.keyfile.Read(keystr)
            if err==nil {
                key+=1
                k.keymap[string(keystr)] = key
            }
        }
    }
    k.keynum = key
    return nil
}


func (k *KeyMap) Get(key string) uint64 {
    return k.keymap[key]
}

//Creates the given key if it doesn't exist
func (k *KeyMap) Create(key string) (uint64, error) {
    //If the key exists, return existing
    if val,ok := k.keymap[key]; ok {
        return val,nil
    }

    //The key does not yet exist.

    //Write the file
    bytestr := []byte(key)
    err := binary.Write(k.keyfile,binary.LittleEndian,uint32(len(bytestr)))
    if (err!=nil) {
        return 0,err
    }
    _,err = k.keyfile.Write(bytestr)
    if (err!=nil) {
        return 0,err
    }

    //Update the keymap
    k.keynum++
    k.keymap[key] = k.keynum

    //The keyfile needs to be flushed
    k.keyfile.Sync()

    return k.keynum,nil

}

//Maps the given key into an integer key. If the key is new, creates a new key and writes it to the keyfile
func (k *KeyMap) Map(key string) (uint64) {
    val,_ := k.Create(key)    //It just so happens that Create does the exact same thing
    return val
}

func OpenKeyMap(fpath string) (*KeyMap,error) {
    if err := MakeDirs(fpath); err!= nil {
        return nil,err
    }


    keyfile,err := os.OpenFile(path.Join(fpath,"keys"), os.O_APPEND|os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        return nil,err
    }

    //Create the KeyMap object with a nil map, since Reload will create and populate it
    k := &KeyMap{keyfile,nil,1}

    err = k.Reload()
    if err!=nil {
        k.Close()
        return nil,err
    }

    return k,nil

}
