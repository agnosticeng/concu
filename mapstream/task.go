package mapstream

import (
	"sort"
)

type Task[T any] struct {
	SequenceNumber uint64
	Value          T
}

type Tasks[T any] []Task[T]

func (ts *Tasks[T]) Insert(job Task[T]) {
	var zero Task[T]

	i := sort.Search(len(*ts), func(i int) bool {
		return (*ts)[i].SequenceNumber > job.SequenceNumber
	})

	*ts = append(*ts, zero)
	copy((*ts)[i+1:], (*ts)[i:])
	(*ts)[i] = job
}
