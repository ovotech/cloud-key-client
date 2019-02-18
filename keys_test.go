package keys

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

var substringTests = []struct {
	in    string
	start string
	end   string
	out   string
}{
	{"hello world", "hello", "", " world"},
	{"hello world", "", "world", "hello "},
	{"hello world", "", "", "hello world"},
	{"", "", "", ""},
}

func TestSubstring(t *testing.T) {
	for _, substringTest := range substringTests {
		substring := subString(substringTest.in, substringTest.start, substringTest.end)
		if substring != substringTest.out {
			t.Errorf("got %q, want %q", substring, substringTest.out)
		}
	}
}

func TestSubStringStartPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	subString("hello world", "panic", "")
}

func TestSubStringEndPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	subString("hello world", "", "panic")
}

func TestMinsSince(t *testing.T) {
	actual := int(minsSince(time.Now()))
	expected := 0
	if int(actual) != expected {
		t.Errorf("Incorrect float returned, got: %d, want: %d.", actual, expected)
	}
}

func TestCheck(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	check(errors.New("this should cause panic"))
}

func TestParseTime(t *testing.T) {
	actual := parseTime(
		time.RFC3339,
		"2012-11-01T23:08:41+00:00").String()
	expected := "2012-11-01 23:08:41 +0000 GMT"
	if expected != actual {
		t.Errorf("Incorrect string returned, got: %s, want: %s.", actual, expected)
	}
}

func TestParseTimePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	parseTime(time.RFC3339, "2012-11-01T25:08:41+00:00")
}

func TestAppendSlice(t *testing.T) {
	sliceOne := make([]Key, 0)
	accountOne := "account-one"
	keyOne := Key{accountOne, "", 0, "", 1, "", Provider{"", ""}}
	sliceOne = append(sliceOne, keyOne)
	sliceTwo := make([]Key, 0)
	accountTwo := "account-two"
	keyTwo := Key{accountTwo, "", 2, "", 3, "", Provider{"", ""}}
	sliceTwo = append(sliceTwo, keyTwo)
	appendedSlice := appendSlice(sliceOne, sliceTwo)

	if !reflect.DeepEqual(appendedSlice[0], keyOne) {
		t.Errorf("Incorrect key returned in slice, got: %+v, want: %+v.",
			appendedSlice[0], keyOne)
	}
	if !reflect.DeepEqual(appendedSlice[1], keyTwo) {
		t.Errorf("Incorrect key returned in slice, got: %+v, want: %+v.",
			appendedSlice[1], keyTwo)
	}
}
