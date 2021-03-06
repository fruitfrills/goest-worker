package goest_worker

import (
	"sync/atomic"
	"context"
)

type prioirityQueue struct {

	// link to forest
	head		*jobHeapNode

	// size of heap
	size 		uint64

	send 		chan jobCall

	receive		chan jobCall

	ctx 		context.Context

	cancel		context.CancelFunc
}

type jobHeapNode struct {

	heap 		*prioirityQueue

	// job for pool
	job jobCall

	// node priority
	priority int

	// parent node with great priority
	parent 		 *jobHeapNode

	// child node with less prioroty
	childHead	 *jobHeapNode

	// one level nodes
	rightSibling *jobHeapNode
}

// This is a factory for creating priority queue
// This implementation use the form of a binomial heap
var PriorityQueue PoolQueue = func (ctx context.Context, capacity int) Queue {
	ctx, cancel := context.WithCancel(ctx)
	queue := &prioirityQueue{
		ctx: ctx,
		cancel: cancel,
		head: nil,
		size: 0,
		send: make(chan jobCall),
		receive: make(chan jobCall, capacity),
	}
	go queue.processor(ctx, queue.send, queue.receive)
	return queue
}

func (queue *prioirityQueue) Insert(job jobCall) {
	select {
	case <-queue.ctx.Done():
		return
	case queue.send <- job:
		return
	}
}

// For the priority queue, the capacity is the size
// of the input channel after which the heap will begin to fill.
func (queue *prioirityQueue) SetCapacity(ctx context.Context, capacity int) {
	queue.cancel()
	queue.ctx, queue.cancel = context.WithCancel(ctx)
	queue.send = make(chan jobCall)
	queue.receive = make(chan jobCall, capacity)
	go queue.processor(queue.ctx, queue.send, queue.receive)
}

func (queue *prioirityQueue) Pop() jobCall {
	select {
	case <- queue.ctx.Done():
		return nil
	case j := <- queue.receive:
		return j
	}
}

func (queue *prioirityQueue) processor(ctx context.Context, send chan jobCall, receive chan jobCall) {

	var exit = func() {
		close(send)
		close(receive)
	}

	go func() {
		for {
			var top *jobHeapNode
			select {
			case <-ctx.Done():
				exit()
				return
			default:
				top = queue.top()
			}

			if top == nil {

				var job jobCall
				select {
				case <-ctx.Done():
					exit()
					return
				case job = <- send:
					break
				}

				// if queue empty try send job to receiver
				select {
				case <-ctx.Done():
					exit()
					return
				case receive <- job:
					continue
				default:
					queue.insertJob(job)
				}
				continue
			}

			select {
			case <-ctx.Done():
				exit()
				return
			case receive <- top.Job():
				top.Remove()
			case value := <- send:
				queue.insertJob(value)
			}

		}
	}()
}

func (jobHeap *prioirityQueue) newJobHeapNode(job jobCall) *jobHeapNode {
	return &jobHeapNode {
		heap: jobHeap,
		job: job,
		parent: nil,
		childHead: nil,
		rightSibling: nil,
		priority: 0,
	}
}

func (bh*prioirityQueue) top() *jobHeapNode {
	if atomic.LoadUint64(&bh.size) == 0 {
		return nil
	}
	return getMinimumNode(bh.head)
}

func (bh *prioirityQueue) remove(node *jobHeapNode) *jobHeapNode {
	removeFromLinkedList(&bh.head, node)

	for _, child := range nodeIterator(node.childHead) {
		removeFromLinkedList(&node.childHead, child)
		bh.insert(child)
	}
	atomic.AddUint64(&bh.size, ^uint64(0))
	return node
}

func (bh *prioirityQueue) insertJob(job jobCall) {
	atomic.AddUint64(&bh.size, 1)
	newnode := bh.newJobHeapNode(job)
	bh.insert(newnode)
}

func (bh *prioirityQueue) Len() uint64 {
	return atomic.LoadUint64(&bh.size)
}

func (bh *prioirityQueue) insert(newnode *jobHeapNode) {
	srnode := getNodeWithPriority(bh.head, newnode.priority)
	if srnode == nil {
		insertIntoLinkedList(&bh.head, newnode)
		return
	}
	removeFromLinkedList(&bh.head, srnode)
	linkednode := linkNodes(srnode, newnode)
	bh.insert(linkednode)

}

func (heap *jobHeapNode) setChild (child *jobHeapNode) {
	insertIntoLinkedList(&heap.childHead, child)
	child.parent = heap
}

func (node *jobHeapNode) Job() (jobCall){
	return node.job
}

func (node *jobHeapNode) Remove() () {
	node.heap.remove(node)
}

func (heap *jobHeapNode) removeChild() {
	removeFromLinkedList(&heap.parent.childHead, heap)
	heap.parent = nil
}

 // linkNodes - create relation between two nodes
func linkNodes(n1 *jobHeapNode, n2 *jobHeapNode) *jobHeapNode {
	if n1.job.Priority() > n2.job.Priority() {
		n1.priority += 1
		n1.setChild(n2)
		return n1
	}
	n2.priority += 1
	n2.setChild(n1)
	return n2
}

func insertIntoLinkedList(head **jobHeapNode, node *jobHeapNode) {
	var prev *jobHeapNode
	var next *jobHeapNode

	prev = nil
	next = *head

	for next != nil && node.priority > next.priority {
		prev = next
		next = next.rightSibling
	}

	if prev == nil && next == nil {
		*head = node
	} else if prev == nil && next != nil {
		node.rightSibling = *head
		*head = node
	} else if prev != nil && next == nil {
		prev.rightSibling = node
	} else if prev != nil && next != nil {
		prev.rightSibling = node
		node.rightSibling = next
	}
}

func removeFromLinkedList(head **jobHeapNode, node *jobHeapNode) {
	leftsib := getLeftsibling(*head, node)
	if leftsib == nil {
		*head = node.rightSibling
	} else {
		leftsib.rightSibling = node.rightSibling
	}
	node.rightSibling = nil
}

func getLeftsibling(head *jobHeapNode, node *jobHeapNode) *jobHeapNode {
	if head == node {
		return nil
	}
	checkNode := head

	for {
		if checkNode.rightSibling == node {
			break
		}
		checkNode = checkNode.rightSibling
	}

	return checkNode
}

func getNodeWithPriority(head *jobHeapNode, priority int) *jobHeapNode {
	checkNode := head
	for {
		if checkNode == nil {
			break
		}
		if checkNode.priority == priority {
			break
		}
		checkNode = checkNode.rightSibling
	}
	return checkNode
}

// getMinimumNode - find job with maximal priority
func getMinimumNode(head *jobHeapNode) *jobHeapNode {

	minnode := head
	checkNode := head.rightSibling

	for {
		if checkNode == nil {
			break
		}
		if checkNode.job.Priority() > minnode.job.Priority() {
			minnode = checkNode
		}
		checkNode = checkNode.rightSibling
	}

	return minnode
}

 // get linked list
func nodeIterator(head *jobHeapNode) [] *jobHeapNode {
	arr := make([] *jobHeapNode, 0, 4)
	rightNode := head
	for {
		if rightNode == nil {
			break
		}
		arr = append(arr, rightNode)
		rightNode = rightNode.rightSibling
	}
	return arr
}