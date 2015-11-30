package tunnel

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

func TestTunnel() {
	m := &bytes.Buffer{}

	ta, err := NewEllipticTunnel(m)
	if err != nil {
		panic(err)
	}

	fmt.Println(ta.Write([]byte("Hello")))
	fmt.Println(m)
	fmt.Println(ta.Write([]byte("World")))
	fmt.Println(m)

	o, _ := ioutil.ReadAll(ta)

	fmt.Println(string(o))
	if string(o) != "HelloWorld" {
		panic(o)
	}
}
