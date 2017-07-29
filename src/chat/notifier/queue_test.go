package notifier

import "testing"

func TestPopPush(t *testing.T) {
	stack := &Queue{}
	item1, item2, item3 := "item1", "item2", "item3"

	stack.Push(item1)
	stack.Push(item2)
	stack.Push(item3)

	item := stack.Pop()

	if item.(string) != item1 {
		t.Fail()
	}

	if len(stack.items) != 2 {
		t.Fail()
	}

	item = stack.Pickup()

	if item.(string) != item2 {
		t.Fail()
	}

	if len(stack.items) != 2 {
		t.Fail()
	}

}
