package jira2pdf

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/karlockhart/go-jira2pdf/pkg/config"
	"github.com/karlockhart/go-jira2pdf/pkg/pdf"
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
		log.Fatalf("Failed to crate jira client: %v", err)
	}

	jiraProjectKeys := viper.GetStringSlice("jira_project_keys")
	if len(jiraProjectKeys) == 0 {
		jiraProjectKeys, err = getAlljiraProjectKeys(jiraClient)
		if err != nil {
			log.Fatalf("Failed to all jira project ids: %v", err)
		}
	}

	projectCount := len(jiraProjectKeys)
	fmt.Printf("Getting ready to build %d PDFs\n", projectCount)

	//Build PDFs for each project with all issues
	for i, projectKey := range jiraProjectKeys {
		jqlQuery := fmt.Sprintf("project = '%s'", projectKey)

		fmt.Println("Fetching issues for " + projectKey)

		issues, err := getIssuesForQuery(jqlQuery, jiraClient)
		if err != nil {
			log.Fatalf("Issue query failed: %v", err)
		}

		err = pdf.BuildPartitionedPDFs(projectKey, issues)
		if err != nil {
			log.Fatalf("Build PDF failed: %v", err)
		}

		fmt.Printf("%d/%d %s complete\n", i+1, projectCount, projectKey)
	}

}

func getAlljiraProjectKeys(jiraClient *jira.Client) ([]string, error) {
	var projects []string
	projectList, _, err := jiraClient.Project.GetList()
	if err != nil {
		return nil, err
	}

	for _, project := range *projectList {
		projects = append(projects, project.Key)
	}

	return projects, nil
}

func getIssuesForQuery(jqlQuery string, jiraClient *jira.Client) ([]jira.Issue, error) {
	var issues []jira.Issue
	pageSize := viper.GetInt("query_page_size")

	for i := 0; ; i += pageSize {
		pagedIssues, _, err := jiraClient.Issue.Search(
			jqlQuery,
			&jira.SearchOptions{
				MaxResults: pageSize,
				StartAt:    i,
				Fields:     getFilteredFields(),
			},
		)
		if err != nil {
			return nil, err
		}

		issues = append(issues, pagedIssues...)
		if len(pagedIssues) < pageSize {
			break
		}
	}

	fmt.Printf("Fetched %d issues\n", len(issues))
	return issues, nil
}

func getFilteredFields() []string {
	issueFilter := []string{}
	issueFields := viper.GetStringSlice("jira_issue_fields")
	for _, issueField := range issueFields {
		issueFilter = append(issueFilter, strings.ToLower(issueField))
	}

	return issueFilter
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
