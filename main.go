package main

import (
	"flag"
)

func main() {
	var opt Opt
	flag.BoolVar(&opt.Install, "i", false, "install and configure tools")
	flag.BoolVar(&opt.Update, "u", false, "update tools if needed")
	flag.Parse()
	New(opt).Run()
}
