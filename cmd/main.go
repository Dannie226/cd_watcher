package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {
	case "deploy":
		os.Exit(deploy())

	case "rollback":
		if len(os.Args) < 3 {
			fmt.Println("Rollback command must have a numeric argument")
			os.Exit(1)
		}

		num, err := strconv.Atoi(os.Args[2])

		if err != nil {
			fmt.Println("Rollback command argument must be a number")
			os.Exit(1)
		}

		os.Exit(rollback(num))
	}
}
