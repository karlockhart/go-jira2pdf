package jira2pdf

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"

	jira "github.com/andygrunwald/go-jira"
	"github.com/jung-kurt/gofpdf"
	"github.com/karlockhart/go-jira2pdf/pkg/config"
	"github.com/spf13/viper"
)

//RunJira2PDF Main entry point for the program
func RunJira2PDF(configFile string) {
	err := config.Load(configFile)
	if err != nil {
		log.Fatal(err)
	}

	jiraClient, err := buildJiraClient()
	if err != nil {
		log.Fatal(err)
	}

	searchOptions := &jira.SearchOptions{
		MaxResults: -1,
	}

	issues, _, err := jiraClient.Issue.Search(viper.GetString("jql_query"), searchOptions)
	if err != nil {
		log.Fatalf("Issue query failed: %v", err)
	}

	// generate PDF
	pdfGenerator := gofpdf.New("P", "mm", "A4", "")
	pdfTR := pdfGenerator.UnicodeTranslatorFromDescriptor("")

	pdfGenerator.AddPage()

	// document title
	pdfGenerator.SetFont("Arial", "B", 16)
	pdfGenerator.SetFillColor(222, 222, 222)
	pdfGenerator.SetTextColor(0, 0, 0)
	pdfGenerator.MultiCell(0, 16, pdfTR("JIRA Issues"), "1", "C", true)

	// document subtitle
	currentPosY := pdfGenerator.GetY()
	pdfGenerator.SetFont("Arial", "", 9)
	pdfGenerator.SetTextColor(0, 0, 0)
	pdfGenerator.MultiCell(0, 9, fmt.Sprintf("Total of issues: %v", len(issues)), "1", "L", false)
	pdfGenerator.SetY(currentPosY)
	pdfGenerator.MultiCell(0, 9, fmt.Sprintf("Created at: %v", time.Now().Format(viper.GetString("datetime_format"))), "1", "R", false)
	pdfGenerator.Ln(4)

	// issues
	for index, issue := range issues {
		// set issue font
		var lineHt float64 = 9
		pdfGenerator.SetFont("Arial", "", lineHt)
		pdfGenerator.SetTextColor(0, 0, 0)

		// parse issue template
		issueText := parseIssueTemplate(index, issue)
		issueText, _ = charmap.Windows1252.NewEncoder().String(issueText)

		html := pdfGenerator.HTMLBasicNew()
		html.Write(lineHt, issueText)

		// draw issue separator
		pdfGenerator.SetDrawColor(195, 195, 195)
		pdfGenerator.Ln(lineHt)
		pdfGenerator.Ln(2)

		pageWidth, _ := pdfGenerator.GetPageSize()
		x, y := pdfGenerator.GetXY()
		marginL, marginR, _, _ := pdfGenerator.GetMargins()
		pdfGenerator.Line(x, y, x+pageWidth-marginR-marginL, y)

		pdfGenerator.Ln(2)
	}

	// save to output filename
	err = pdfGenerator.OutputFileAndClose("Test.pdf")

	if err != nil {
		log.Fatalf("Erro while save PDF: %v", err)
	}

}

func buildJiraClient() (*jira.Client, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}

	jiraClient, err := jira.NewClient(client, viper.GetString("jira_url"))

	if err != nil {
		return nil, err
	}

	jiraClient.Authentication.SetBasicAuth(viper.GetString("J2P_USERNAME"), viper.GetString("J2P_PASSWORD"))

	return jiraClient, nil
}

func parseIssueTemplate(index int, issue jira.Issue) string {
	issueText := viper.GetString("jira_issue_template")

	issueText = strings.Replace(issueText, "[issue.key]", issue.Key, -1)
	issueText = strings.Replace(issueText, "[issue.id]", issue.ID, -1)
	issueText = strings.Replace(issueText, "[issue.fields.description]", issue.Fields.Description, -1)
	issueText = strings.Replace(issueText, "[issue.fields.duedate]", issue.Fields.Duedate, -1)
	issueText = strings.Replace(issueText, "[issue.fields.expand]", issue.Fields.Expand, -1)
	issueText = strings.Replace(issueText, "[issue.fields.resolutiondate]", issue.Fields.Resolutiondate, -1)
	issueText = strings.Replace(issueText, "[issue.fields.summary]", issue.Fields.Summary, -1)
	issueText = strings.Replace(issueText, "[issue.fields.timeestimate]", strconv.Itoa(issue.Fields.TimeEstimate), -1)
	issueText = strings.Replace(issueText, "[issue.fields.timeoriginalestimate]", strconv.Itoa(issue.Fields.TimeOriginalEstimate), -1)
	issueText = strings.Replace(issueText, "[issue.fields.timespent]", strconv.Itoa(issue.Fields.TimeSpent), -1)

	issueText = strings.Replace(issueText, "[issue.fields.project.name]", issue.Fields.Project.Name, -1)
	issueText = strings.Replace(issueText, "[issue.fields.project.description]", issue.Fields.Project.Description, -1)
	issueText = strings.Replace(issueText, "[issue.fields.project.id]", issue.Fields.Project.ID, -1)
	issueText = strings.Replace(issueText, "[issue.fields.project.key]", issue.Fields.Project.Key, -1)

	issueText = strings.Replace(issueText, "[issue.fields.type.name]", issue.Fields.Type.Name, -1)
	issueText = strings.Replace(issueText, "[issue.fields.type.description]", issue.Fields.Type.Description, -1)
	issueText = strings.Replace(issueText, "[issue.fields.type.id]", issue.Fields.Type.ID, -1)

	if issue.Fields.Priority != nil {
		issueText = strings.Replace(issueText, "[issue.fields.priority.id]", issue.Fields.Priority.ID, -1)
		issueText = strings.Replace(issueText, "[issue.fields.priority.name]", issue.Fields.Priority.Name, -1)
	} else {
		issueText = strings.Replace(issueText, "[issue.fields.priority.id]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.priority.name]", "", -1)
	}

	if issue.Fields.AggregateProgress != nil {
		issueText = strings.Replace(issueText, "[issue.fields.aggregateprogress.progress]", strconv.Itoa(issue.Fields.AggregateProgress.Progress), -1)
		issueText = strings.Replace(issueText, "[issue.fields.aggregateprogress.total]", strconv.Itoa(issue.Fields.AggregateProgress.Total), -1)
	} else {
		issueText = strings.Replace(issueText, "[issue.fields.aggregateprogress.progress]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.aggregateprogress.total]", "", -1)
	}

	if issue.Fields.Progress != nil {
		issueText = strings.Replace(issueText, "[issue.fields.progress.progress]", strconv.Itoa(issue.Fields.Progress.Progress), -1)
		issueText = strings.Replace(issueText, "[issue.fields.progress.total]", strconv.Itoa(issue.Fields.Progress.Total), -1)
	} else {
		issueText = strings.Replace(issueText, "[issue.fields.progress.progress]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.progress.total]", "", -1)
	}

	if issue.Fields.Assignee != nil {
		issueText = strings.Replace(issueText, "[issue.fields.assignee.name]", issue.Fields.Assignee.Name, -1)
		issueText = strings.Replace(issueText, "[issue.fields.assignee.emailaddrress]", issue.Fields.Assignee.EmailAddress, -1)
		issueText = strings.Replace(issueText, "[issue.fields.assignee.displayname]", issue.Fields.Assignee.DisplayName, -1)
		issueText = strings.Replace(issueText, "[issue.fields.assignee.key]", issue.Fields.Assignee.Key, -1)
	} else {
		issueText = strings.Replace(issueText, "[issue.fields.assignee.name]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.assignee.emailaddrress]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.assignee.displayname]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.assignee.key]", "", -1)
	}

	if issue.Fields.Creator != nil {
		issueText = strings.Replace(issueText, "[issue.fields.creator.name]", issue.Fields.Creator.Name, -1)
		issueText = strings.Replace(issueText, "[issue.fields.creator.emailaddrress]", issue.Fields.Creator.EmailAddress, -1)
		issueText = strings.Replace(issueText, "[issue.fields.creator.displayname]", issue.Fields.Creator.DisplayName, -1)
		issueText = strings.Replace(issueText, "[issue.fields.creator.key]", issue.Fields.Creator.Key, -1)
	} else {
		issueText = strings.Replace(issueText, "[issue.fields.creator.name]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.creator.emailaddrress]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.creator.displayname]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.creator.key]", "", -1)
	}

	if issue.Fields.Reporter != nil {
		issueText = strings.Replace(issueText, "[issue.fields.reporter.name]", issue.Fields.Reporter.Name, -1)
		issueText = strings.Replace(issueText, "[issue.fields.reporter.emailaddrress]", issue.Fields.Reporter.EmailAddress, -1)
		issueText = strings.Replace(issueText, "[issue.fields.reporter.displayname]", issue.Fields.Reporter.DisplayName, -1)
		issueText = strings.Replace(issueText, "[issue.fields.reporter.key]", issue.Fields.Reporter.Key, -1)
	} else {
		issueText = strings.Replace(issueText, "[issue.fields.reporter.name]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.reporter.emailaddrress]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.reporter.displayname]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.reporter.key]", "", -1)
	}

	if issue.Fields.Status != nil {
		issueText = strings.Replace(issueText, "[issue.fields.status.name]", issue.Fields.Status.Name, -1)
		issueText = strings.Replace(issueText, "[issue.fields.status.description]", issue.Fields.Status.Description, -1)
		issueText = strings.Replace(issueText, "[issue.fields.status.id]", issue.Fields.Status.ID, -1)
	} else {
		issueText = strings.Replace(issueText, "[issue.fields.status.name]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.status.description]", "", -1)
		issueText = strings.Replace(issueText, "[issue.fields.status.id]", "", -1)
	}

	if len(issue.Fields.Created) > 0 {
		dateTime, err := time.Parse(viper.GetString("api_datetime_format"), issue.Fields.Created)

		if err == nil {
			issueText = strings.Replace(issueText, "[issue.fields.created]", dateTime.Format(viper.GetString("datetime_format")), -1)
		} else {
			log.Printf("Error on parse field \"created\"! %v\n", err)

			issueText = strings.Replace(issueText, "[issue.fields.created]", "", -1)
		}
	}

	return issueText
}
