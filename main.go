package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	docopt "github.com/docopt/docopt.go"
)

func loadRomZip(fn string) []uint8 {
	r, err := zip.OpenReader(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".gb") {
			rc, err := f.Open()
			if err != nil {
				log.Fatal(err)
			}
			defer rc.Close()
			buf, err := ioutil.ReadAll(rc)
			if err != nil {
				log.Fatal(err)
			}
			r := make([]uint8, len(buf))
			for i, b := range buf {
				r[i] = uint8(b)
			}
			return r
		}
	}
	return []uint8{}
}

func loadRom(fn string) []uint8 {
	if strings.HasSuffix(fn, ".zip") {
		return loadRomZip(fn)
	}
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	r := make([]uint8, len(buf))
	for i, b := range buf {
		r[i] = uint8(b)
	}
	return r
}

func main() {
	doc := `usage: go-gboy <rom>`
	args, _ := docopt.Parse(doc, nil, true, "", false)

	rom := loadRom(args["<rom>"].(string))

	mc := newMemoryController(rom)
	c := newCpu(mc, nil)
	fmt.Println(c)
	for {
		c.loop()
		fmt.Println(c)
		if len(commandTable[c.inst[0]].String()) == 0 {
			panic("unknown opcode")
		}
	}
}
