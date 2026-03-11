package main

type DynamicArray[T any] struct {
	data []T
	len  int
	cap  int
}

func NewDynamicArray[T any]() *DynamicArray[T] {
	return &DynamicArray[T]{
		data: nil,
		len:  0,
		cap:  0,
	}
}

func NewDynamicArrayWithLen[T any](length int) *DynamicArray[T] {
	s := make([]T, length)
	return &DynamicArray[T]{
		data: s,
		len:  length,
		cap:  length,
	}
}

func NewDynamicArrayWithLenCap[T any](length, capacity int) *DynamicArray[T] {
	s := make([]T, capacity)
	return &DynamicArray[T]{
		data: s,
		len:  length,
		cap:  capacity,
	}
}

func (da *DynamicArray[T]) Append(val ...T) *DynamicArray[T] {
	newDA := *da
	for _, v := range val {
		if newDA.len == newDA.cap {
			newCap := newDA.cap * 2
			if newDA.cap == 0 {
				newCap = 1
			} else if newDA.cap >= 1024 {
				newCap = newDA.cap + newDA.cap/4
			}

			newData := make([]T, newCap)
			if newDA.data != nil {
				copy(newData, newDA.data[:newDA.len])
			}
			newDA.data = newData
			newDA.cap = newCap
		}
		newDA.data[newDA.len] = v
		newDA.len++
	}
	return &newDA
}

func (da *DynamicArray[T]) Get(index int) (T, bool) {
	if index < 0 || index >= da.len {
		return *new(T), false
	}
	return da.data[index], true
}

func (da *DynamicArray[T]) Set(index int, val T) bool {
	if index < 0 || index >= da.len {
		return false
	}
	da.data[index] = val
	return true
}
