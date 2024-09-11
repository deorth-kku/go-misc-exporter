package common

import "time"

type float interface {
	~float32 | ~float64
}

func FloatDuration[T float](f T) time.Duration {
	return time.Duration(T(time.Second) * f)
}
