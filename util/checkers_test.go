package util_test

import (
	"testing"
	. "github.com/naaman/busl/util"
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
