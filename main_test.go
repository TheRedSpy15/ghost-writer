package main

import (
	"testing"
)

// test functions to make the program handles intentionally bad parameters correctly
func TestBadParameters(t *testing.T) {
	getCoinValuesTimeRange(0, "")
	getCoinValuesTimeRange(1, "")
	getCoinValuesTimeRange(1, "bad")

	getValueFromPostgres("")
	getValueFromPostgres("bad")
}
