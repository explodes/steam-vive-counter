package main

import "time"

func b2i(v bool) int {
	if v {
		return 1
	}
	return 0
}

func i2b(i int) bool {
	return i == 0
}

func i2t(i int64) time.Time {
	return time.Unix(i, 0)
}

func t2i(t time.Time) int64 {
	return t.Unix()
}
