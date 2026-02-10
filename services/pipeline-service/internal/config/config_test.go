package config

import (
	"reflect"
	"testing"
)

func TestSplitAndTrimSplitsAndTrimsValues(t *testing.T) {
	input := " repo1 ,repo2,  repo3 ,, "
	want := []string{"repo1", "repo2", "repo3"}

	got := splitAndTrim(input, ",")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitAndTrim(%q) = %#v, want %#v", input, got, want)
	}
}

