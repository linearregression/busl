package util

import (
	. "gopkg.in/check.v1"
	"reflect"
)

// -----------------------------------------------------------------------
// IsTrue checker.
type isTrueChecker struct {
	*CheckerInfo
}

// The IsTrue checker tests whether the obtained value is true.
//
// For example:
//
//    c.Assert(err, IsTrue)
//
var IsTrue Checker = &isTrueChecker{
	&CheckerInfo{Name: "IsTrue", Params: []string{"value"}},
}

func (checker *isTrueChecker) Check(params []interface{}, names []string) (result bool, error string) {
	return isTrue(params[0]), ""
}

func isTrue(obtained interface{}) (result bool) {
	if obtained == nil {
		result = false
	} else {
		switch v := reflect.ValueOf(obtained); v.Kind() {
		case reflect.Bool:
			return v.Bool() == true
		}
	}
	return false
}
