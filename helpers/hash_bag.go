package helpers

import (
	"redun-pendancy/utils"
	"sort"
)

type HashBag[T comparable] struct {
	items map[T]int
}

func NewHashBag[T comparable]() *HashBag[T] {
	return &HashBag[T]{
		items: make(map[T]int),
	}
}

func (hashBag *HashBag[T]) Add(item T) int {
	newCount := hashBag.items[item] + 1
	hashBag.items[item] = newCount
	return newCount
}

func (hashBag *HashBag[T]) GetItemCount(item T) int {
	return hashBag.items[item]
}

func (hashBag *HashBag[T]) GetEntriesCount() int {
	return len(hashBag.items)
}

func (hashBag *HashBag[T]) GetTotalCount() int {
	total := 0
	for _, count := range hashBag.items {
		total += count
	}
	return total
}

func (hashBag *HashBag[T]) GetItems() map[T]int {
	return hashBag.items
}

func (hashBag *HashBag[T]) GetAsEntries() []utils.MapEntry[T, int] {
	return utils.GetMapEntries(hashBag.items)
}

func (hashBag *HashBag[T]) GetSortedEntriesAsc() []utils.MapEntry[T, int] {
	entries := hashBag.GetAsEntries()
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Value < entries[j].Value
	})
	return entries
}

func (hashBag *HashBag[T]) GetSortedEntriesDesc() []utils.MapEntry[T, int] {
	entries := hashBag.GetAsEntries()
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Value > entries[j].Value
	})
	return entries
}

func (hashBag *HashBag[T]) Remove(item T) int {
	count, exists := hashBag.items[item]
	if !exists {
		return 0
	}

	if count == 1 {
		delete(hashBag.items, item)
		return 0
	}

	newCount := count - 1
	hashBag.items[item] = newCount
	return newCount
}
