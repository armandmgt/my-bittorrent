package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [inputfile]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if flag.NArg() < 1 {
		fmt.Println("Input file is missing.")
		os.Exit(1)
	}

	tf, err := ReadTorrent(args[0])
	if err != nil {
		fmt.Println("Not a valid torrent file.")
		os.Exit(1)
	}
	fmt.Println(tf)
	if err := DownloadTorrent(tf); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
