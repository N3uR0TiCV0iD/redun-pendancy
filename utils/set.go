package utils

type Set[T comparable] map[T]bool

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func (set Set[T]) Add(item T) bool {
	if set.Contains(item) {
		return false
	}
	set[item] = true
	return true
}

func (set Set[T]) UnionWith(other Set[T]) int {
	lastCount := len(set)
	for item := range other {
		set.Add(item)
	}
	return lastCount - len(set)
}

func (set Set[T]) AddRange(items []T) int {
	lastCount := len(set)
	for _, item := range items {
		set.Add(item)
	}
	return lastCount - len(set)
}

func (set Set[T]) Contains(item T) bool {
	_, exists := set[item]
	return exists
}

func (set Set[T]) Count() int {
	return len(set)
}

func (set Set[T]) Remove(item T) bool {
	if !set.Contains(item) {
		return false
	}
	delete(set, item)
	return true
}

func (set Set[T]) IntersectWith(other Set[T]) int {
	lastCount := len(set)
	DeleteIf(set, func(item T, _ bool) bool {
		return !other.Contains(item)
	})
	return lastCount - len(set)
}

func (set Set[T]) IsEmpty() bool {
	return len(set) == 0
}

func (set Set[T]) Clear() {
	for key := range set {
		delete(set, key)
	}
}
