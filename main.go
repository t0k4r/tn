package main

import (
	"flag"
	"tn/tn"
)

func main() {
	var opt tn.Opt
	flag.BoolVar(&opt.Install, "i", false, "install and configure tools")
	flag.BoolVar(&opt.Update, "u", false, "update tools if needed")
	flag.Parse()
	tn.New(opt).Run()
}
