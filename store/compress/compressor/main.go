package main

import (
	"fmt"
	"io"
	"os"
)

func openFiles(from, to string) (*os.File, *os.File, error) {
	fdFrom, err := os.Open(from)
	if err != nil {
		return nil, nil, err
	}

	fdTo, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		fdFrom.Close()
		return nil, nil, err
	}

	return fdFrom, fdTo, nil
}

func main() {
	// TODO: flags package nutzen!!!!
	// TODO: ..und am besten gleich in format-util einbauen.
	// TODO: ..dieses package dann nicht als `main` sondern
	//       als das was momentan `compress` ist.
	if os.Args[3] == "c" {

		from, to, err := openFiles(os.Args[1], os.Args[2])
		if err != nil {
			fmt.Println(err)
		}

		sw := NewWriter(to, AlgoSnappy)

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
