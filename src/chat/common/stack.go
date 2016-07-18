package common

//FifoStack implements fifo stack (first in, first out)
type FifoStack struct {
	items []interface{}
}

//Push adds item to stack
func (f *FifoStack) Push(item interface{}) {
	f.items = append(f.items, item)
}

//Pop removes and returns item from stack
func (f *FifoStack) Pop() (item interface{}) {
	if len(f.items) == 0 {
		return
	}

	item = f.items[0]
	f.items = f.items[1:]
	return
}

//Pickup  returns an item without removing
func (f *FifoStack) Pickup() (item interface{}) {
	if len(f.items) == 0 {
		return
	}

	item = f.items[0]
	return
}

//Len return stack length
func (f *FifoStack) Len() int {
	return len(f.items)
}
