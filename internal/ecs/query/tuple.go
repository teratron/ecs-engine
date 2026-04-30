package query

// Tuple2 is the value yielded by [Query2.All]. Go's range-over-func iterators
// expose at most two values per step (key, value), so multi-component queries
// pack their component pointers into a tuple struct that travels as the
// "value" half of the [iter.Seq2].
type Tuple2[A, B any] struct {
	A *A
	B *B
}

// Tuple3 is the value yielded by [Query3.All]. See [Tuple2] for the rationale.
type Tuple3[A, B, C any] struct {
	A *A
	B *B
	C *C
}
