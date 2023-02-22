package collection

type Iterator[V any] struct {
	slice  []V
	length int
	pos    int
}

func NewIterator[V any](in ...V) *Iterator[V] {
	return &Iterator[V]{
		slice:  in,
		length: len(in),
		pos:    -1,
	}
}

func (iter *Iterator[V]) Next() (V, bool) {
	if iter.pos+1 < iter.length {
		iter.pos++
		return iter.slice[iter.pos], true
	}

	var zero V
	return zero, false
}

func (iter *Iterator[V]) Peek() (V, bool) {
	if iter.pos+1 < iter.length {
		return iter.slice[iter.pos+1], true
	}

	var zero V
	return zero, false
}
