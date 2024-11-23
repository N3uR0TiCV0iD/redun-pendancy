package utils

import (
	"fmt"
	"sort"
)

type MapEntry[K comparable, V any] struct {
	Key   K
	Value V
}

func Map[TSource any, TResult any](source []TSource, mappingFunc func(TSource) TResult) []TResult {
	result := make([]TResult, len(source))
	for index, item := range source {
		result[index] = mappingFunc(item)
	}
	return result
}

func FirstOrDefault[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	var defaultValue T
	return defaultValue, false
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func GroupBy[T any, K comparable](items []T, keySelector func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, item := range items {
		key := keySelector(item)
		groups[key] = append(groups[key], item)
	}
	return groups
}

func GetMapKeys[K comparable, V any](source map[K]V) []K {
	keys := make([]K, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	return keys
}

func GetMapEntries[K comparable, V any](source map[K]V) []MapEntry[K, V] {
	entries := make([]MapEntry[K, V], 0, len(source))
	for key, value := range source {
		entry := MapEntry[K, V]{
			Key:   key,
			Value: value,
		}
		entries = append(entries, entry)
	}
	return entries
}

func IndexOf[T any](slice []T, startIndex int, predicate func(T) bool) int {
	PanicIfOutOfBounds(startIndex, len(slice))
	for index := startIndex; index < len(slice); index++ {
		if predicate(slice[index]) {
			return index
		}
	}
	return -1
}

func LastIndexOf[T any](slice []T, startIndex int, predicate func(T) bool) int {
	PanicIfOutOfBounds(startIndex, len(slice))
	for index := startIndex; index >= 0; index-- {
		if predicate(slice[index]) {
			return index
		}
	}
	return -1
}

func InsertAt[T any](slice []T, index int, element T) []T {
	if index == len(slice) {
		return append(slice, element)
	}

	PanicIfOutOfBounds(index, len(slice))
	result := make([]T, len(slice)+1)

	copy(result, slice[:index])
	result[index] = element
	copy(result[index+1:], slice[index:])

	return result
}

func RemoveAt[T any](slice []T, index int) []T {
	PanicIfOutOfBounds(index, len(slice))
	return append(slice[:index], slice[index+1:]...)
}

func RemoveIf[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if !predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func DeleteIf[K comparable, V any](items map[K]V, predicate func(K, V) bool) {
	for key, value := range items {
		if predicate(key, value) {
			delete(items, key)
		}
	}
}

func SortIntsDesc(slice []int) {
	ascSorter := sort.IntSlice(slice)
	descSorter := sort.Reverse(ascSorter)
	sort.Sort(descSorter)
}

func PanicIfOutOfBounds(index, length int) {
	if index < 0 || index >= length {
		panic(fmt.Sprintf("index %d out of bounds [0:%d]", index, length))
	}
}

func IndexOutOfBoundsError(index, length int) error {
	return fmt.Errorf("index %d out of bounds [0:%d]", index, length)
}
