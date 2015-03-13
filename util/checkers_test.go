package util_test

import (
	check "github.com/heroku/busl/Godeps/_workspace/src/gopkg.in/check.v1"
	. "github.com/heroku/busl/util"
	"testing"
)

type checkerTest struct {
	check    check.Checker
	input    []interface{}
	expected bool
}

var checkerTests = []checkerTest{
	checkerTest{IsTrue, []interface{}{true}, true},
	checkerTest{IsTrue, []interface{}{false}, false},
	checkerTest{IsFalse, []interface{}{false}, true},
	checkerTest{IsFalse, []interface{}{true}, false},
	checkerTest{IsEmptyString, []interface{}{""}, true},
	checkerTest{IsEmptyString, []interface{}{"d"}, false},
	checkerTest{IsEmptyString, []interface{}{nil}, false},
}

func TestCheckers(t *testing.T) {
	for _, c := range checkerTests {
		actual, _ := c.check.Check(c.input, []string{})
		if actual != c.expected {
			t.Errorf("Expected %T to return %v, but got %v.", c.check, c.expected, actual)
		}
	}
}
