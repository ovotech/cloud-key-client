package keys

import (
	"errors"
	"reflect"
	"testing"
)

var substringTests = []struct {
	in    string
	start string
	end   string
	out   string
	err   error
}{
	{"hello world", "hello", "", " world", nil},
	{"hello world", "", "world", "hello ", nil},
	{"hello world", "", "", "hello world", nil},
	{"", "", "", "", nil},
	{"hello world", "should_produce_error", "", "", errors.New("")},
}

func TestSubstring(t *testing.T) {
	for _, substringTest := range substringTests {
		substr, err := subString(substringTest.in, substringTest.start, substringTest.end)
		if (err != nil && substringTest.err == nil) || (err == nil && substringTest.err != nil) {
			t.Errorf("got an unexpected number of errors")
		} else if substr != substringTest.out {
			t.Errorf("got %q, want %q", substr, substringTest.out)
		}
	}
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
