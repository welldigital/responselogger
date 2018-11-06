package stats

import (
	"sort"
)

// Value allows stats to be collected.
type Value interface {
	Value() int
}

// Sum of the items.
func Sum(items []Value) (op int) {
	for _, item := range items {
		op += item.Value()
	}
	return
}

// Average returns the top count items.
func Average(items []Value) float64 {
	if len(items) == 0 {
		return 0
	}
	return float64(Sum(items)) / float64(len(items))
}

// Top returns the top count items.
func Top(items []Value, count int) []Value {
	if len(items) <= count || count == 0 {
		return items
	}
	values := valueSlice(items)
	sort.Sort(values)
	return values[len(values)-count:]
}

// Bottom returns the bottom count items.
func Bottom(items []Value, count int) []Value {
	if len(items) <= count || count == 0 {
		return items
	}
	values := valueSlice(items)
	sort.Sort(sort.Reverse(values))
	return values[0:count]
}

type valueSlice []Value

func (p valueSlice) Len() int           { return len(p) }
func (p valueSlice) Less(i, j int) bool { return p[i].Value() < p[j].Value() }
func (p valueSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
