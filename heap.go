package goest_worker

import (
	"sync/atomic"
)

type jobHeap struct {

	// link to forest
	head		 *jobHeapNode

	// size of heap
	size 			uint64
}

type jobHeapNode struct {

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

func newJobHeap() *jobHeap {
	return &jobHeap {
		head: nil,
		size: 0,
	}
}

func newJobHeapNode(job jobCall) *jobHeapNode {
	return &jobHeapNode {
		job: job,
		parent: nil,
		childHead: nil,
		rightSibling: nil,
		priority: 0,
	}
}

func (bh* jobHeap) Top() *jobHeapNode {
	if atomic.LoadUint64(&bh.size) == 0 {
		return nil
	}
	node := getMinimumNode(bh.head)
	return node
}

func (bh *jobHeap) Remove(node *jobHeapNode) *jobHeapNode {
	removeFromLinkedList(&bh.head, node)

	for _, child := range nodeIterator(node.childHead) {
		removeFromLinkedList(&node.childHead, child)
		bh.insert(child)
	}
	atomic.AddUint64(&bh.size, ^uint64(0))
	return node
}

func (bh* jobHeap) Insert(job jobCall) {
	atomic.AddUint64(&bh.size, 1)
	newnode := newJobHeapNode(job)
	bh.insert(newnode)
}

func (bh* jobHeap) Len() uint64 {
	return atomic.LoadUint64(&bh.size)
}


func (bh *jobHeap) insert(newnode *jobHeapNode) {
	srnode := getNodeWithPriority(bh.head, newnode.priority)

	if srnode == nil {
		insertIntoLinkedList(&bh.head, newnode)
	} else {
		removeFromLinkedList(&bh.head, srnode)
		linkednode := linkNodes(srnode, newnode)
		bh.insert(linkednode)
	}
}



func (heap *jobHeapNode) setChild (child *jobHeapNode) {
	insertIntoLinkedList(&heap.childHead, child)
	child.parent = heap
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
	} else {
		n2.priority += 1
		n2.setChild(n1)
		return n2
	}
}

func insertIntoLinkedList(head * *jobHeapNode, node *jobHeapNode) {
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