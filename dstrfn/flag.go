package dstrfn

import "flag"

var debug bool

func init() {
	flag.BoolVar(&debug, "dstrfn.debug", false, "Debug mode?")
}
