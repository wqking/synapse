package main

import (
	"flag"
	"os"
	"strings"

	"github.com/phoreproject/synapse/p2p"
	"github.com/phoreproject/synapse/utils"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/phoreproject/synapse/explorer"
)

func main() {
	utils.CheckNTP()

	config := explorer.LoadConfig()
	changed, newLimit, err := utils.ManageFdLimit()
	if err != nil {
		panic(err)
	}
	if changed {
		logrus.Infof("changed ulimit to: %d", newLimit)
	}

	ex, err := explorer.NewExplorer(config)
	if err != nil {
		panic(err)
	}

	err = ex.StartExplorer()
	if err != nil {
		panic(err)
	}
}
