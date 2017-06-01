package jira2pdf

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

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

		fmt.Printf("%d/%d Getting issues for %s.pdf\n", i+1, projectCount, projectKey)
		issues, _, err := jiraClient.Issue.Search(
			jqlQuery,
			&jira.SearchOptions{MaxResults: -1},
		)
		if err != nil {
			log.Fatalf("Issue query failed: %v", err)
		}

		fmt.Printf("%d/%d Building %s.pdf...\n", i+1, projectCount, projectKey)
		err = pdf.Build(projectKey, issues)
		if err != nil {
			log.Fatalf("Build PDF failed: %v", err)
		}
		fmt.Printf("%d/%d %s.pdf complete\n", i+1, projectCount, projectKey)
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
