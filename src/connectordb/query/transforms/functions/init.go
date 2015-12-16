/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package functions

//init does the necessary registration of all the builtin functions
func init() {
	average.Register()
	sum.Register()
	changed.Register()
}
