package main

import (
	"consistent-hashing/consistenthash"
	"fmt"
	"sync"
)

func main(){
	ring := consistenthash.New(100,nil)
	nodes := []consistenthash.StorageNode{
		{Name: "A", Host: "239.67.52.72"},
		{Name: "B", Host: "137.70.131.229"},
		{Name: "C", Host: "98.5.87.182"},
		{Name: "D", Host: "11.225.158.95"},
		{Name: "E", Host: "203.187.116.210"},
	}
	for _,n := range nodes {
		if err := ring.AddNode(n); err != nil{
			panic(err)
		}
	}
	files := []string{"f1.txt","f2.txt","f3.txt","f4.txt","f5.txt"}
	fmt.Println("--before scaling--")
	for _,f := range files{
		n,err := ring.Get(f)
		if err != nil {
			panic(err)
		}
		fmt.Printf("file %s resides on node %s \n",f,n.(consistenthash.StorageNode).Name)
	}

	before := make(map[string]string)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("file-%d.txt", i)
		n, _ := ring.Get(key)
		before[key] = n.(consistenthash.StorageNode).Name
	}



	newNodes := []consistenthash.StorageNode{
		{Name: "F", Host: "107.117.238.203"},
		{Name: "G", Host: "27.161.219.131"},
	}
	for _, n := range newNodes {
		if err := ring.AddNode(n); err != nil {
			panic(err)
		}
	}

	moved := 0
	for key, oldNode := range before {
		n, _ := ring.Get(key)
		if n.(consistenthash.StorageNode).Name != oldNode {
			moved++
		}
	}
	fmt.Printf("keys that moved after adding 2 nodes: %d / 100 \n", moved)

	fmt.Println("\n-- after adding f and g")
	for _,f := range files{
		n,err := ring.Get(f)
		if err != nil {
			panic(err)
		}
		fmt.Printf("file %s resides on node %s\n",f,n.(consistenthash.StorageNode).Name)
	}
	
	fmt.Println("\n concurrent access smoke test")
	var wg sync.WaitGroup
	for i := 0;i<50;i++{
		wg.Add(1)
		go func(i int){
			defer wg.Done()
			key := fmt.Sprintf("concurrent-key-%d",i)
			if _,err := ring.Get(key); err != nil{
				panic(err)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("all concurrent read completed without a race")
}