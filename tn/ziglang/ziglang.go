package ziglang

import "fmt"

type Zig struct {
}

func New() Zig {
	return Zig{}
}

func (z Zig) Install() {
	fmt.Println("zig install")
}
func (z Zig) Update() {
	fmt.Println("zig update")
}
