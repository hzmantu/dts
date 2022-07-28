package main

import (
	"code.hzmantu.com/dts/cmd"
	"log"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cmd.Execute()
}
