package consistenthash

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	ErrEmptyRing = errors.New("consistenthash: ring is empty")
	ErrNodeExists = errors.New("consistenthash: node already exists on the ring")
	ErrNodeNotFound = errors.New("consistenthash: node not found on ring")
)

type Ring struct {
	mu sync.RWMutex
	hash Hashfunc
	replicas int //no of virtual points per real node
	sortedKeys []uint32
	pointToID map[uint32]string //ring position -> nodeid
	nodes 	  map[string]Node // nodeid -> node
}

func New(replicas int,fn Hashfunc) *Ring{
	if fn == nil {
		fn = defaultHash
	}
	if replicas <= 0 {
		replicas = 1
	}
	return &Ring{
		hash: fn,
		replicas: replicas,
		pointToID: make(map[uint32]string),
		nodes: make(map[string]Node), //make before use
	}
}

func virtualKey(id string,i int)string {
	return fmt.Sprintf("%s#%d",id,i)
}


func (r *Ring) AddNode(n Node)error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	id := n.ID()
	if _,ok := r.nodes[id]; ok {
		return fmt.Errorf("%w: %s",ErrNodeExists,id)
	}

	for i:=0; i < r.replicas;i++ {
		point := r.hash(virtualKey(id,i))
		if _,exists := r.pointToID[point]; !exists{
			r.sortedKeys = append(r.sortedKeys,point)
		}
		r.pointToID[point] = id
	}
	//Sort all the hash positions from smallest to largest, 
	//store this node in the node map using its ID
	sort.Slice(r.sortedKeys,func(i,j int) bool {
		return r.sortedKeys[i] < r.sortedKeys[j]
	})
	r.nodes[id] = n
	return nil
}


func (r *Ring) RemoveNode(n Node)error{
	r.mu.Lock()
	defer r.mu.Unlock()

	id := n.ID()
	if _,ok := r.nodes[id]; !ok{
		return fmt.Errorf("%w: %s",ErrNodeNotFound,id)
	}
	//removing node and its virtual nodes as well
	for i := 0;i<r.replicas;i++{
		point := r.hash(virtualKey(id,i))
		delete(r.pointToID,point)
	}
	//rebuild sorted keys from pointToID
	newKeys := make([]uint32,0,len(r.pointToID))
	for k := range r.pointToID{
		newKeys = append(newKeys,k)
	}
	sort.Slice(newKeys,func(i,j int) bool {
		return newKeys[i] < newKeys[j]
	})
	r.sortedKeys = newKeys
	delete(r.nodes,id)
	return nil
}

func (r *Ring) Get(key string)(Node,error){
	r.mu.RLock() //read lock
	defer r.mu.RUnlock()

	if len(r.sortedKeys) == 0{
		return nil,ErrEmptyRing
	}
	h := r.hash(key)
	idx := sort.Search(len(r.sortedKeys),func(i int) bool {
		return r.sortedKeys[i] >= h
	})
	if idx == len(r.sortedKeys){
		idx = 0 //wrapping (ring)
	}
	id := r.pointToID[r.sortedKeys[idx]]
	return r.nodes[id],nil
}

func (r *Ring) Nodes() []string{
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string,0,len(r.nodes))
	for id := range r.nodes{
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}