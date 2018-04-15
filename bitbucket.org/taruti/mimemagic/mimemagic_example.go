// +build ignore

package main

import "fmt"
import "bitbucket.org/taruti/mimemagic"
import "os"

func main() {
	b := make([]byte, 1024)
	for _,fn := range os.Args {
		f,e := os.Open(fn)
		if e!=nil { panic(e) }
		f.Read(b)
		fmt.Printf("%-30s %s\n", mimemagic.Match("",b),fn)
		f.Close()
	}
}
