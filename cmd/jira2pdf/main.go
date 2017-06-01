package main

import (
	"github.com/docopt/docopt-go"
	"github.com/karlockhart/go-jira2pdf/pkg/jira2pdf"
)

func main() {
	usage := `Jira2PDF

Usage:
  jira2pdf (-f=<file>)
  jira2pdf -h | --help
  jira2pdf --version

Options:
  -f=<file> --file  YAML config file for creating the pdf.
  -h --help         Show this screen.
  -v --version      Show version.`

	args, _ := docopt.Parse(usage, nil, true, "Jira2PDF 1.0", false)

	jira2pdf.RunJira2PDF(args["--file"].(string))
}
