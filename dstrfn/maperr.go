package dstrfn

import (
	"fmt"
	"sort"
)

type MapError struct {
	Master error
	Tasks  map[int]error
	Len    int
}

func (err MapError) Error() string {
	if err.Master != nil {
		return fmt.Sprintf("%v: tasks failed %d/%d", err.Master, len(err.Tasks), err.Len)
	}
	return fmt.Sprintf("tasks failed %d/%d", len(err.Tasks), err.Len)
}

func keys(tasks map[int]error) []int {
	if len(tasks) == 0 {
		return nil
	}
	ks := make([]int, 0, len(tasks))
	for k := range tasks {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	return ks
}
