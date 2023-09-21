package tn

import (
	"log"
	"tn/tn/golang"
)

type Opt struct {
	Install bool
	Update  bool
}

type tn struct {
	opt     Opt
	update  []func() error
	install []func() error
}

func (t tn) Run() {
	if t.opt.Install {
		for _, installer := range t.install {
			installer()
		}
	} else if t.opt.Update {
		for _, updater := range t.update {
			updater()
		}
	}
}

func New(opt Opt) tn {
	if !opt.Install && !opt.Update {
		log.Fatal("nothing to do")
	}
	return tn{
		opt:     opt,
		update:  []func() error{golang.Update},
		install: []func() error{golang.Install},
	}
}
