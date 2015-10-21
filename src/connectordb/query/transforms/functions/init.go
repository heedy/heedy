package functions

//init does the necessary registration of all the builtin functions
func init() {
	average.Register()
	sum.Register()
	changed.Register()
}
