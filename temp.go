package main

import "fmt"

var (
	stuff = map[string]func(*Store, string) (error){
		"neat": (*Store).neat,
	}
)

type Store struct {}

func main() {
	store := &Store{}
	stuff["neat"](store, "wow!")
	i := 0
	for ; i < 10; i++ {
		fmt.Println(i)
	}
	fmt.Println(i, i < 10)
	fmt.Println("p.kek"[2:])
}

func (store *Store) neat(arg string) error {
	fmt.Println(arg)
	return nil
}