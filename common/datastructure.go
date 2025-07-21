package common

type Queue[T any] interface {
    Enqueue(item T)           // Add an item to the queue
    Dequeue() (T, bool)       // Remove and return the front item (false if empty)
    IsEmpty() bool            // Check if the queue is empty
}


type fifoQueue[T any] struct {
    items []T
}

func (q *fifoQueue[T]) Enqueue(item T) {
    q.items = append(q.items, item)
}

func (q *fifoQueue[T]) Dequeue() (T, bool) {
    if len(q.items) == 0 {
        var zero T
        return zero, false
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item, true
}

func (q *fifoQueue[T]) IsEmpty() bool {
    return len(q.items) == 0
}

