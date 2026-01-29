package main

// SortAlgorithm represents the sorting algorithm to use in the experiment.
type SortAlgorithm int

const (
	QuickSort SortAlgorithm = iota
	RadixSort
)

func (s SortAlgorithm) String() string {
	switch s {
	case QuickSort:
		return "QuickSort"
	case RadixSort:
		return "RadixSort"
	default:
		return "Unknown"
	}
}

// DataPattern represents the data pattern to use as a noise factor in the experiment.
type DataPattern int

const (
	Random DataPattern = iota
	Sorted
	ReverseSorted
	ManyDuplicates
	NearlySorted
)

func (d DataPattern) String() string {
	switch d {
	case Random:
		return "Random"
	case Sorted:
		return "Sorted"
	case ReverseSorted:
		return "ReverseSorted"
	case ManyDuplicates:
		return "ManyDuplicates"
	case NearlySorted:
		return "NearlySorted"
	default:
		return "Unknown"
	}
}
