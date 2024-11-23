package helpers

import "redun-pendancy/utils"

type OrderedMap[K comparable, V any] struct {
	order       []K
	items       map[K]V
	itemIndices map[K]int
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		items:       make(map[K]V),
		itemIndices: make(map[K]int),
	}
}

func (orderMap *OrderedMap[K, V]) Get(key K) (V, bool) {
	value, exists := orderMap.items[key]
	return value, exists
}

func (orderMap *OrderedMap[K, V]) Set(key K, value V) {
	_, exists := orderMap.items[key]
	if !exists {
		orderMap.order = append(orderMap.order, key)
		orderMap.itemIndices[key] = len(orderMap.items)
	}
	orderMap.items[key] = value
}

func (orderMap *OrderedMap[K, V]) GetAt(index int) V {
	key := orderMap.order[index]
	return orderMap.items[key]
}

func (orderMap *OrderedMap[K, V]) GetIndex(key K) (int, bool) {
	index, exists := orderMap.itemIndices[key]
	return index, exists
}

func (orderMap *OrderedMap[K, V]) GetCount() int {
	return len(orderMap.items)
}

func (orderMap *OrderedMap[K, V]) GetOrderedKeys() []K {
	return orderMap.order
}

func (orderMap *OrderedMap[K, V]) GetOrderedValues() []V {
	values := make([]V, 0, len(orderMap.order))
	for _, key := range orderMap.order {
		values = append(values, orderMap.items[key])
	}
	return values
}

func (orderMap *OrderedMap[K, V]) SwapItems(from K, to K) bool {
	if from == to {
		return false
	}

	fromIndex, exists := orderMap.GetIndex(from)
	if !exists {
		return false
	}

	toIndex, exists := orderMap.GetIndex(to)
	if !exists {
		return false
	}

	orderMap.SwapIndices(fromIndex, toIndex)
	return true
}

func (orderMap *OrderedMap[K, V]) SwapIndices(from int, to int) {
	if from == to {
		return
	}
	backup := orderMap.order[to]
	orderMap.order[to] = orderMap.order[from]
	orderMap.order[from] = backup
}

func (orderMap *OrderedMap[K, V]) MoveItem(item K, newIndex int) bool {
	utils.PanicIfOutOfBounds(newIndex, len(orderMap.order))
	oldIndex, exists := orderMap.GetIndex(item)
	if !exists {
		return false
	}

	if oldIndex == newIndex {
		return false
	}
	orderMap.MoveIndex(oldIndex, newIndex)
	return true
}

func (orderMap *OrderedMap[K, V]) MoveIndex(from int, to int) {
	if from == to {
		return
	}
	order := orderMap.order
	item := order[from]
	order = utils.RemoveAt(order, from)

	//Adjust "to" if it comes after "from" (as items have shifted due to removal)
	if to > from {
		to--
	}
	order = utils.InsertAt(order, to, item)
	orderMap.order = order
}

func (orderMap *OrderedMap[K, V]) Remove(key K) bool {
	_, exists := orderMap.items[key]
	if !exists {
		return false
	}

	removeIndex := orderMap.itemIndices[key]
	order := utils.RemoveAt(orderMap.order, removeIndex)
	delete(orderMap.itemIndices, key)
	delete(orderMap.items, key)
	orderMap.order = order

	// Update indices of remaining keys
	for index := removeIndex; index < len(order); index++ {
		key := order[index]
		orderMap.itemIndices[key] = index
	}
	return true
}
