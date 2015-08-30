package create

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	dirPermission  = 0775
	filePermission = 0664
)

var (
	flags       = flag.NewFlagSet("create command", flag.ExitOnError)
	commandName = "create"

	// types
	web bool

	// languages
	golang bool
)

func Command(args []string) {
	parseFlags(args)

	onlyOne([]bool{golang}, "language")
	onlyOne([]bool{web}, "type")

	name := fetchName()

	createGoWebService(name, golang, web)
}

func createGoWebService(name string, golang bool, web bool) {
	if !(golang && web) {
		return
	}

	os.Mkdir(name, dirPermission)
	downloadFile("https://github.com/nanoservice/template.web.go/raw/master/main.go", name+"/main.go")
	downloadFile("https://github.com/nanoservice/template.web.go/raw/master/README.md", name+"/README.md")
}

func downloadFile(url string, path string) {
	out, err := os.Create(path)
	if err != nil {
		downloadError(url, err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		downloadError(url, err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		downloadError(url, err)
	}
}

func downloadError(url string, err error) {
	log.Fatalf("Unable to download file %s: %v", url, err)
}

func fetchName() string {
	args := flags.Args()
	if len(args) != 1 {
		fmt.Printf("Expected argument NAME, but got %d arguments\n", len(args))
		flags.Usage()
	}

	return args[0]
}

func onlyOne(flags []bool, name string) {
	count := 0
	for _, flag := range flags {
		if flag {
			count += 1
		}
	}

	if count != 1 {
		log.Fatalf("Only one %s should have been chosen, but got: %d", name, count)
	}
}

func parseFlags(args []string) {
	flags.BoolVar(&web, "web", false, "Enables web template")
	flags.BoolVar(&golang, "go", false, "Enables go template")

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s %s OPTIONS NAME:\n", os.Args[0], commandName)
		flags.PrintDefaults()
		os.Exit(1)
	}

	flags.Parse(args)
}
