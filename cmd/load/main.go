package main

import (
	"encoding/json"
	"fmt"
	"github.com/serge-v/stv/channel/smithsonian"
)

type Item = smithsonian.Item

type Response struct {
	Count int
	List  []Item
}

func testSerialization() {
	resp := Response{
		Count: 1,
		List:  []Item{Item{Name: "test", Href: "testhref"}},
	}

	buf, err := json.Marshal(&resp)
	if err != nil {
		panic(err)
	}
	println(string(buf))

	var respcli Response
	if err := json.Unmarshal(buf, &respcli); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", respcli)
}

func main() {
	testSerialization()
	list, err := smithsonian.GetVideos()
	if err != nil {
		panic(err)
	}
	resp := Response{
		Count: len(list),
		List:  list,
	}
	fmt.Printf("%+v\n", resp)
}
