package main

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/namsral/flag"
	gogitlab "github.com/plouc/go-gitlab-client"
)

type context struct {
	Gitlab *gogitlab.Gitlab
	Match  string
}

func getContext(args []string) context {
	fs := flag.NewFlagSetWithEnvPrefix(args[0], "GLAB", flag.PanicOnError)

	host := fs.String("host", "", "name of the gitlab host")
	apiPath := fs.String("apipath", "", "api path on the gitlab host")
	token := fs.String("token", "", "token for gitlab")
	match := fs.String("r", "", "regular expression to match the projects")

	if err := fs.Parse(args[1:]); err != nil {
		log.Fatalf("Error parsing project options: %s\n", err)
	}

	return context{
		Gitlab: gogitlab.NewGitlab(*host, *apiPath, *token),
		Match:  *match}
}

func filterProjects(projects []*gogitlab.Project, match string) []gogitlab.Project {
	var fp []gogitlab.Project
	r, _ := regexp.Compile(fmt.Sprintf("(?i)%s", match))

	for _, project := range projects {
		if r.MatchString(project.Path) {
			fp = append(fp, *project)
		}
	}
	return fp
}

func listProjects(args []string) {
	c := getContext(args)
	projects, err := c.Gitlab.Projects()
	if err != nil {
		log.Fatalf("Error fetching projects: %s\n", err)
	}

	fp := filterProjects(projects, c.Match)

	fmt.Printf("\n\tprojects found: %d\n\n", len(fp))
	for c, project := range fp {
		fmt.Printf("%d. %s: %s\n", (c + 1), project.Name, project.PathWithNamespace)
		fmt.Printf("  > %s\n\n", project.SshRepoUrl)
	}
	fmt.Println(" ")
}

func main() {

	subcommand := ""
	var args []string
	if len(os.Args) > 1 {
		subcommand = os.Args[1]
		args = append(os.Args[:1], os.Args[2:]...)
	}

	switch subcommand {
	case "projects":
		listProjects(args)
	// case "replay":
	default:
		log.Fatal("Available commands: projects")
	}

}
