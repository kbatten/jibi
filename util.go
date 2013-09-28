package main

import (
	"archive/zip"
	"io/ioutil"
	"log"
	"strings"
)

func bytesToWord(msb, lsb uint8) uint16 {
	return uint16(msb)<<8 + uint16(lsb)
}

func wordToBytes(w uint16) (lsb uint8, msb uint8) {
	return uint8(w >> 8), uint8(w & 0xFF)
}

func bytesToAddress(msb, lsb uint8) address {
	return address(uint16(msb)<<8 + uint16(lsb))
}

func readRomZipFile(fn string) []uint8 {
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

func readRomFile(fn string) []uint8 {
	if strings.HasSuffix(fn, ".zip") {
		return readRomZipFile(fn)
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
