package formatting

import (
	"testing"
)

type row struct {
	Foo string
	Bar string
	Baz int
}

func TestTable(t *testing.T) {
	table := NewTable(
		[]row{
			row{"foo", "bar", 3},
			row{"foofoo", "barbar", 33}},
		[]string{"Bar", "Baz", "Foo"})

	expected := "Bar     Baz  Foo   \nbar     3    foo   \nbarbar  33   foofoo"

	if got := table.Format(false); got != expected {
		t.Errorf("got %q instead of expected %q", got, expected)
	}

	table2 := NewTable(
		[]*row{
			&row{"foo", "bar", 3},
			&row{"foofoo", "barbar", 33}},
		[]string{"Bar", "Baz", "Foo"})
	if got := table2.Format(false); got != expected {
		t.Errorf("got %q instead of expected %q", got, expected)
	}
}
