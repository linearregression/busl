package util_test

import (
	. "github.com/naaman/busl/util"
	"testing"
)

func TestIsTrueTrueValueIsTrue(t *testing.T) {
	trueCheck, _ := IsTrue.Check([]interface{}{true}, []string{})
	if !trueCheck {
		t.Errorf("Expected IsTrue to return true, but got false.")
	}
}

func TestIsTrueFalseValueIsFalse(t *testing.T) {
	trueCheck, _ := IsTrue.Check([]interface{}{false}, []string{})
	if trueCheck {
		t.Errorf("Expected IsTrue to return false, but got true.")
	}
}

func TestIsFalseFalseValueIsTrue(t *testing.T) {
	falseCheck, _ := IsFalse.Check([]interface{}{false}, []string{})
	if !falseCheck {
		t.Errorf("Expected IsFalse to return true, but got false.")
	}
}

func TestIsFalseTrueValueIsFalse(t *testing.T) {
	falseCheck, _ := IsFalse.Check([]interface{}{true}, []string{})
	if falseCheck {
		t.Errorf("Expected IsFalse to return false, but got true.")
	}
}

func TestIsEmptyStringEmptyStringValueIsTrue(t *testing.T) {
	emptyStringCheck, _ := IsEmptyString.Check([]interface{}{""}, []string{})
	if !emptyStringCheck {
		t.Errorf("Expected IsEmptyString to return true, but got false.")
	}
}

func TestIsEmptyStringStringWithDataIsFalse(t *testing.T) {
	emptyStringCheck, _ := IsEmptyString.Check([]interface{}{"d"}, []string{})
	if emptyStringCheck {
		t.Errorf("Expected IsEmptyString to return true, but got false.")
	}
}

func TestIsEmptyStringNilValueIsFalse(t *testing.T) {
	emptyStringCheck, _ := IsEmptyString.Check([]interface{}{nil}, []string{})
	if emptyStringCheck {
		t.Errorf("Expected IsEmptyString to return true, but got false.")
	}
}
