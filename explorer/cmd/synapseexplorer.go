package main

import (
	"github.com/phoreproject/synapse/utils"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/phoreproject/synapse/explorer"
)

func main() {
	utils.CheckNTP()

	config := explorer.LoadConfig()

	ex, err := explorer.NewExplorer(config)
	if err != nil {
		panic(err)
	}

	err = ex.StartExplorer()
	if err != nil {
		panic(err)
	}
}
