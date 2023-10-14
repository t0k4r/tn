package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"tn/golang"
	"tn/ziglang"

	"golang.org/x/sync/errgroup"
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
	g := new(errgroup.Group)
	if t.opt.Install {
		for _, installer := range t.install {
			g.Go(installer)
		}
	} else if t.opt.Update {
		for _, updater := range t.update {
			g.Go(updater)
		}
	}
	err := g.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func New(opt Opt) tn {
	setup()
	if !opt.Install && !opt.Update {
		log.Fatal("nothing to do")
	}
	if opt.Install && opt.Update {
		log.Fatal("chose one")
	}
	return tn{
		opt:     opt,
		update:  []func() error{golang.Update, ziglang.Update},
		install: []func() error{golang.Install, ziglang.Install},
	}
}

func setup() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	bash_profile := fmt.Sprintf("%v/.bash_profile", home)
	content, err := os.ReadFile(bash_profile)
	if err != nil {
		log.Fatal(err)
	}
	setup := ". ~/.tn/setup"
	if !strings.Contains(string(content), setup) {
		file, err := os.OpenFile(bash_profile, os.O_APPEND|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		_, err = file.WriteString(fmt.Sprintf("%v\n", setup))
		if err != nil {
			log.Fatal(err)
		}
	}
	dir := fmt.Sprintf("%v/.tn", home)
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	setupFile := fmt.Sprintf("%v/.tn/setup", home)

	file, err := os.Create(setupFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.WriteString("export PATH=$PATH:~/.tn/zig\nexport PATH=$PATH:~/.tn/go/bin\n")
	if err != nil {
		log.Fatal(err)
	}
}
