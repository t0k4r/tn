package main

import (
	"flag"
)

func main() {
	var opt Opt
	flag.BoolVar(&opt.Install, "i", false, "install and setup path")
	flag.BoolVar(&opt.Update, "u", false, "update if needed")
	flag.Parse()
	New(opt).Run()
}
