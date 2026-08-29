package main

import "fmt"

func usage() {
	fmt.Println(`Usage: cd_watcher COMMAND

Commands:

deploy
Deploys the bundles in the deploy directory

rollback [number]
Rolls back the bundles, the passed in number of times removing both
release directories and database entries`)
}
