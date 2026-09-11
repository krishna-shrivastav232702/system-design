package consistenthash

type Node interface {
	ID() string
}

type StorageNode struct {
	Name string
	Host string
}

func (n StorageNode) ID() string { //func belongs to storageNode (n is storagenode type) and a receiver
	return n.Host
}