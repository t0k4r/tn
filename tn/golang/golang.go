package golang

import "fmt"

type Go struct {
}

func New() Go {
	return Go{}
}

func (g Go) Install() {
	fmt.Println("go install")
}
func (g Go) Update() {
	fmt.Println("go update")
}
