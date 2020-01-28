package main

import (
	"fmt"

	"Gtkmmcargo/builder"
)

func main() {
	b := builder.New("/home/piotr/Projects/Gtkmm/Test/")
	b.AddFile("test.cc")
	//builder.PrintGtkmmFlags()
	ok := b.Build("testapp", true)
	if !ok {
		fmt.Println("failed")
	}
}
