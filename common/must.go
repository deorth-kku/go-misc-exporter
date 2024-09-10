package common

func Must0(err error) {
	if err != nil {
		panic(err)
	}
}

func Must[T any](arg T, err error) T {
	if err != nil {
		panic(err)
	}
	return arg
}

func Must2[T any, U any](arg1 T, arg2 U, err error) (T, U) {
	if err != nil {
		panic(err)
	}
	return arg1, arg2
}

func Must3[T any, U any, V any](arg1 T, arg2 U, arg3 V, err error) (T, U, V) {
	if err != nil {
		panic(err)
	}
	return arg1, arg2, arg3
}
