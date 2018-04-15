// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type section struct {
	Name string
	matchers []matcher
}
type matcher struct {
	Level, Offset, Range int
	Value []byte
	Mask  []byte
}

func main() {
	inp,e := ioutil.ReadFile("/usr/share/mime/magic")
	if e!=nil { panic(e) }
	if !bytes.Equal([]byte("MIME-Magic\x00\n"),inp[0:12]) { panic("Invalid file header") }
	inp = inp[12:]
	out := new(bytes.Buffer)
	var s section
	ss := []section{}
	i  := 0
	duplicates := map[string]bool{}
	for len(inp)>0 {
		inp,s = pSection(inp, out)
		if duplicates[s.Name] { continue }
		duplicates[s.Name] = true
		ss = append(ss,s)
		fmt.Fprintf(out, "%#v: &sarr[%d],\n",s.Name,i)
		i++
	}
	
	fmt.Printf(`package mimemagic

var smap = map[string]*section{
%s}

var sarr = [...]section{
`,out.Bytes())
	for i:=0; i<len(ss);i++ {
		x := fmt.Sprintf("%#v,\n",ss[i])
		os.Stdout.WriteString(strings.Replace(x,"main.","",-1))
	}
	os.Stdout.WriteString("}\n\n")
}

func pSection(b []byte, out *bytes.Buffer) ([]byte,section) {
	if b[0] != '[' { panic("Expected [") }
	i := bytes.IndexByte(b,':')
	b = b[i+1:]
	j := bytes.IndexByte(b,']')
	if b[j+1]!='\n' { panic("Expected newline") }
	name := string(b[:j])
	
	b    = b[j+2:]
	mss := []matcher{}
	var m matcher
	for {
		b,m = pMatcher(name,b)
		mss = append(mss, m)
		if len(b)==0 || b[0]=='[' { break }
	}

	return b,section{name,mss}
}

func pMatcher(name string, b []byte) ([]byte,matcher) {
	var lvl,soff,rang int
	var mask []byte
	lvl,b = pInt(b)
	if b[0] != '>' { panic("Expecting '>'") }
	b = b[1:]
	soff,b = pInt(b)
	if b[0] != '=' { panic("Expecting '='") }
	dlen := int(b[2]) + int(b[1])
	b = b[3:]
	value  := b[0:dlen]
	b = b[dlen:]

	for {
		switch b[0] {
		case '\n':
			return b[1:],matcher{lvl,soff,rang,value,mask}
		case '&':
			mask = b[1:1+len(value)]
			b = b[1+len(value):]
		case '~':
			fmt.Fprintf(os.Stderr,"%s: Byte shifting not supported - using default\n",name)
			_,b = pInt(b[1:])
		case '+':
			rang,b = pInt(b[1:])
		default:
			panic("Who are you? '"+string(b[0:1])+"'\n")
		}
	}
	panic("not reached")
}



func pInt(b []byte) (int,[]byte) {
	acc := 0
	for len(b)>0 {
		switch b[0] {
		case '1': acc = 1 + (acc*10)
		case '2': acc = 2 + (acc*10)
		case '3': acc = 3 + (acc*10)
		case '4': acc = 4 + (acc*10)
		case '5': acc = 5 + (acc*10)
		case '6': acc = 6 + (acc*10)
		case '7': acc = 7 + (acc*10)
		case '8': acc = 8 + (acc*10)
		case '9': acc = 9 + (acc*10)
		case '0': acc = 0 + (acc*10)
		default:  return acc,b
		}
		b = b[1:]
	}
	return acc,b
}
