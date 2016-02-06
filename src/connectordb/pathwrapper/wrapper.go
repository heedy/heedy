package pathwrapper

import "connectordb/operator"

// Wrapper takes the underlying database, and an Operator instance, and enables path-based operations on the
// operator. The reason it can't use just the operator is due to permissions values - while for things liek AuthOperator,
// returned objects are only if permissions are given, the PathOperatorMixin needs to query the database for the information
// needed to actually create the desired query in the first place!
//
// Implementation warning: Only the ID portion of the returned user/device/stream is guaranteed to be valid, as other fields,
// including name might have been censored by previous operators (such as authoperator). Therefore functions in Wrapper
// can't rely on anything other than ID in returned users/devices/streams. Such implementations also can't call anything OTHER
// than read functions (other than their mirror function in the underlying operator) due to possible permissions errors.
type Wrapper struct {
	operator.Operator
}

// Wrap wraps an Operator such that it conforms the the PathOperator interface
func Wrap(o operator.Operator) Wrapper {
	return Wrapper{o}
}
