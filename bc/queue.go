package bc

import "sync"

type Queue struct {
	elements []interface{}
	lock     sync.Mutex
}

func NewQueue() *Queue {
	q := new(Queue)
	q.elements = make([]interface{}, 0)
	return q
}

func (q *Queue) Enqueue(e interface{}) *Queue {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.elements = append(q.elements, e)
	return q
}

func (q *Queue) Dequeue() (e interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()
	e, q.elements = q.elements[0], q.elements[1:len(q.elements)]
	return
}

func (q *Queue) Size() int {
	return len(q.elements)
}
