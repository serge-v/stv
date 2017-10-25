package main

//go:generate go run embed.go

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var list = []string{
	"generic.html",
	"main.html",
	"play.html",
}

func main() {
	f, err := os.Create("../stv_embed.go")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(f, "package main")

	repl := strings.NewReplacer("css", "CSS", "html", "HTML", ".", "")

	for _, item := range list {
		text, err := ioutil.ReadFile(item)
		if err != nil {
			panic(err)
		}

		varName := repl.Replace(item)

		fmt.Fprintln(f, "\nconst (")
		fmt.Fprintf(f, "	%s = `%s`\n", varName, string(text))
		fmt.Fprintln(f, ")")
	}

	if err := f.Close(); err != nil {
		panic(err)
	}
}
