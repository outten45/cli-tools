package main

import (
	"fmt"
	"log"
	"os"

	"github.com/namsral/flag"
	gogitlab "github.com/plouc/go-gitlab-client"
)

func projectList(args []string) {
	fs := flag.NewFlagSetWithEnvPrefix(args[0], "GLAB", flag.PanicOnError)

	host := fs.String("host", "", "name of the gitlab host")
	apiPath := fs.String("apipath", "", "api path on the gitlab host")
	token := fs.String("token", "", "token for gitlab")

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Errorf("error: %s\n", err)
	}

	gitlab := gogitlab.NewGitlab(*host, *apiPath, *token)
	fmt.Printf("gitlab: %+v \n", *gitlab)

}

func main() {

	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	} else {
		fmt.Println("Used needs to include a subcommand.")
		return
	}

	args := append(os.Args[:1], os.Args[2:]...)

	switch mode {
	case "projects":
		projectList(args)
	// case "replay":
	default:
		log.Fatal("Available commands: projects")
	}

}
