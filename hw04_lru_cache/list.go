package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	firstNode *ListItem
	lastNode  *ListItem
	len       int
}

func NewList() List {
	return new(list)
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.firstNode
}

func (l *list) Back() *ListItem {
	return l.lastNode
}

func (l *list) PushFront(v interface{}) *ListItem {
	newNode := ListItem{v, nil, nil}

	if l.firstNode == nil {
		l.firstNode = &newNode
		l.lastNode = &newNode
	} else {
		l.insertBefore(l.firstNode, &newNode)
	}
	l.len++
	return &newNode
}

func (l *list) PushBack(v interface{}) *ListItem {
	if l.lastNode == nil {
		return l.PushFront(v)
	}

	newNode := ListItem{v, nil, nil}
	l.insertAfter(l.lastNode, &newNode)
	l.len++

	return &newNode
}

func (l *list) Remove(i *ListItem) {
	if i.Prev == nil {
		l.firstNode = i.Next
	} else {
		i.Prev.Next = i.Next
	}

	if i.Next == nil {
		l.lastNode = i.Prev
	} else {
		i.Next.Prev = i.Prev
	}
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	l.Remove(i)
	l.PushFront(i.Value)
}

func (l *list) insertBefore(node *ListItem, newNode *ListItem) {
	newNode.Next = node
	if node.Prev == nil {
		newNode.Prev = nil
		l.firstNode = newNode
	} else {
		newNode.Prev = node.Prev
		node.Prev.Next = newNode
	}
	node.Prev = newNode
}

func (l *list) insertAfter(node *ListItem, newNode *ListItem) {
	newNode.Prev = node
	if node.Next == nil {
		newNode.Next = nil
		l.lastNode = newNode
	} else {
		newNode.Next = node.Next
		node.Next.Prev = newNode
	}
	node.Next = newNode
}
