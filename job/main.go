package main

import (
	"github.com/horizoncd/horizon/job/cmd"
)

func main() {
	cmd.Run(cmd.ParseFlags())
}
