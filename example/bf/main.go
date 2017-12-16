package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yanolab/gopjit"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " FILE")
		return
	}
	buf, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	for i := range buf {
		switch buf[i] {
		case '>':
			fmt.Fprintf(&b, "p++\n")
		case '<':
			fmt.Fprintf(&b, "p--\n")
		case '+':
			fmt.Fprintf(&b, "b[p]++\n")
		case '-':
			fmt.Fprintf(&b, "b[p]--\n")
		case ',':
		case '.':
			fmt.Fprintf(&b, "os.Stdout.Write(b[p:p+1])\n")
		case '[':
			fmt.Fprintf(&b, "for b[p] != 0 {\n")
		case ']':
			fmt.Fprintf(&b, "}\n")
		}
	}

	src := fmt.Sprintf(`package main
		import "os"
	func F0() {
		b := make([]byte, 30000)
		p := 0
		%s
	}`, b.String())

	jit := gopjit.NewJIT()
	sym, err := jit.BuildSrc(src)
	if err != nil {
		panic(err)
	}

	(sym.(func()))()
}
