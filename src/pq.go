package main

import "container/heap"

type PQItem struct{
    Priority float64
    Serial int
    State State
}

type PriorityQueue []PQItem

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i,j int) bool{
    if pq[i].Priority==pq[j].Priority { return pq[i].Serial < pq[j].Serial }
    return pq[i].Priority < pq[j].Priority
}
func (pq PriorityQueue) Swap(i,j int){ pq[i],pq[j]=pq[j],pq[i] }
func (pq *PriorityQueue) Push(x interface{}){ *pq = append(*pq, x.(PQItem)) }
func (pq *PriorityQueue) Pop() interface{}{
    old := *pq
    n := len(old)
    it := old[n-1]
    *pq = old[:n-1]
    return it
}

func (pq *PriorityQueue) PushItem(item PQItem){ heap.Push(pq, item) }
func (pq *PriorityQueue) PopItem() PQItem{ return heap.Pop(pq).(PQItem) }
