package ttbase

func RejectRandom(weight int) bool {

	//TODO: exponential? 1000 * 2^(-0.06894 * weight)
	return false // don't reject
}
