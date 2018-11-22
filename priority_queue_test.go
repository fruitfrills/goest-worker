package goest_worker

import (
	"testing"
	"context"
)

func TestPrioirityQueue_Insert(t *testing.T) {
	var mock = &mockJobCall{}
	pq := PriorityQueue(context.TODO(), 8)
	pq.Insert(mock)
	job := pq.Pop()
	if job == nil || job.(*mockJobCall) != mock {
		t.Error("wrong job")
	}
}

func TestPrioirityQueue_Insert2(t *testing.T) {
	var mock = &mockJobCall{}
	var mock2 = &mockJobCall{
		priority: 1,
	}
	var mock3 = &mockJobCall{
		priority: 2,
	}

	pq := PriorityQueue(context.TODO(), 0)
	pq.Insert(mock)
	pq.Insert(mock2)
	pq.Insert(mock3)
	job := pq.Pop()
	if job == nil || job.(*mockJobCall) != mock3 {
		t.Error("wrong prioirty job")
	}
	job = pq.Pop()
	if job == nil || job.(*mockJobCall) != mock2 {
		t.Error("wrong prioirty job")
	}
	job = pq.Pop()
	if job == nil || job.(*mockJobCall) != mock {
		t.Error("wrong prioirty job")
	}
}

func TestPrioirityQueue_Insert3(t *testing.T) {
	pq := PriorityQueue(context.TODO(), 0)
	for i := 0; i <= 100000; i++ {
		pq.Insert(&mockJobCall{
			priority: i,
		})
	}
	for i := 100000; i >= 0; i-- {
		job := pq.Pop()
		if job.Priority() != i {
			t.Error("incorrect priority")
		}
	}
}

func TestPrioirityQueue_Len(t *testing.T) {
	var mock= &mockJobCall{}
	pq := PriorityQueue(context.TODO(), 0)
	pq.Insert(mock)
	pq.Insert(mock)
	pq.Insert(mock)
	if pq.Len() != 3 {
		t.Error("incorect length")
	}
}

