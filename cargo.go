package main

import (
	"fmt"

	"Gtkmmcargo/builder"
)

func main() {
	b := builder.New("/home/piotr/Projects/Gtkmm/Test/")
	b.AddFile("test.cc")
	//builder.PrintGtkmmFlags()
	b.Compile()
	ok, outString, errString := b.Link("testapp")
	if !ok {
		fmt.Println("Out:", outString)
		fmt.Println("Err:", errString)
	}
}
