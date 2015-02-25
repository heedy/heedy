Dtypes
=================

This module handles all of the datatypes of the database, and provides a simple wrapper over TimeBatchDB which automatically correctly converts stuff to the byte arrays TimeBatchDB expects.

Types are of the form: `a/b/c`.

The `a` is the underlying storage type of the object. So for example, `s/html` is an HTML webpage - but to store it, we only need the "s" (string) part.

A good example of this is `f[2]/gps`. There are special array types in the database. GPS coordinates are stored as 2 floating point numbers (lat,long). The dtype core only uses the f[2] to store stuff.

Lastly, there are "max length" limits. For example, a text message with a 160 character limit (max 160 char) can be encoded as follows: `s[-160]/sms`. That's right - negative values are interpreted as "up to". The 0 value is interpreted as "unlimited", so `f[0]` is an array of unlimited (within reason!) length.

Supported Types
-------

- x - binary byte array
- s - text string
- f - float64
- i - int64
- b - bool
- x\[] - byte array with given length limits
- s\[] - string with given length limits
- f\[]
- i\[]

Usage
-------

The types are encoded in the obvious way: `BinaryType` is a binary datapoint type, and `BinaryDatapoint` is the actual struct holding data.

To get the types and data:
```go
dtype,ok := GetType("s[-160]/sms")
if !ok {
    panic(0)
}
dpoint := dtype.New()

jsonthing.Unmarshal(&dpoint)

if !dtype.IsValid(&dpoint) {
    log.Printf("Unmarshalled data does not satisfy constraints of 160 chars max")
}
```
