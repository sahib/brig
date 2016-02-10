package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if os.Args[3] == "c" {

		from, to, err := openFiles(os.Args[1], os.Args[2])
		if err != nil {
			fmt.Println(err)
		}

		sw := NewWriter(to)
		io.Copy(sw, from)
		//if err := to.Close(); err != nil {
		//	fmt.Println(err)
		//}
		sw.Close()

	} else if os.Args[3] == "d" {

		from, to, err := openFiles(os.Args[1], os.Args[2])
		if err != nil {
			fmt.Println(err)
		}
		sr := NewReader(from)
		io.Copy(to, sr)
		to.Close()

	} else {

		fmt.Println("no compress or decompress flag.")
	}
}
