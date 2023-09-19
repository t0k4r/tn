package tn

import (
	"log"
	"tn/tn/golang"
	"tn/tn/ziglang"
)

type Opt struct {
	Install bool
	Update  bool
}

type TN interface {
	Install()
	Update()
}

type tn struct {
	opt   Opt
	tools []TN
}

func (t tn) Run() {
	if t.opt.Install {
		for _, tool := range t.tools {
			tool.Install()
		}
	}
	if t.opt.Update {
		for _, tool := range t.tools {
			tool.Update()
		}
	}
}

func New(opt Opt) tn {
	if !opt.Install && !opt.Update {
		log.Fatal("nothing to do")
	}
	var tools []TN
	tools = append(tools, ziglang.New())
	tools = append(tools, golang.New())
	return tn{opt, tools}
}
