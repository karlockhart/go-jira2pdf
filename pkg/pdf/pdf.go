package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/text/encoding/charmap"

	jira "github.com/andygrunwald/go-jira"
	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/viper"
)

//BuildPartitionedPDFs builds one or more pdfs for a project
func BuildPartitionedPDFs(project string, issues []jira.Issue) error {
	issuesPerPDF := viper.GetInt("issues_per_pdf")
	partitionMap := getParitionMap(len(issues), issuesPerPDF)

	for i, partition := range partitionMap {
		err := Build(fmt.Sprintf("%s_%d", project, i+1), project, issues[partition[0]:partition[1]])
		if err != nil {
			return err
		}
	}

	return nil
}

//Generates a parition map for golang slices [[start1,end1],[start2,end2]]
//these items can then be used to extract a sub-slice e.g. slice[start1:end1]
func getParitionMap(collectionSize int, partitionSize int) [][]int {
	partitionMap := [][]int{}

	for partitionBegin := 0; partitionBegin < collectionSize; partitionBegin += partitionSize {
		paritionEnd := partitionBegin + partitionSize
		if paritionEnd > collectionSize {
			paritionEnd = collectionSize
		}
		partitionMap = append(partitionMap, []int{partitionBegin, paritionEnd})
	}

	return partitionMap
}

//Build a pdf from jira issues
func Build(fileName string, title string, issues []jira.Issue) error {
	pdfGenerator := gofpdf.New("P", "mm", "A4", "")
	pdfTR := pdfGenerator.UnicodeTranslatorFromDescriptor("")

	pdfGenerator.AddPage()

	// document title
	pdfGenerator.SetFont("Arial", "B", 16)
	pdfGenerator.SetFillColor(222, 222, 222)
	pdfGenerator.SetTextColor(0, 0, 0)
	pdfGenerator.MultiCell(0, 16, pdfTR(title+" Issues"), "1", "C", true)

	// document subtitle
	currentPosY := pdfGenerator.GetY()
	pdfGenerator.SetFont("Arial", "", 9)
	pdfGenerator.SetTextColor(0, 0, 0)
	pdfGenerator.MultiCell(0, 9, fmt.Sprintf("Total of issues: %v", len(issues)), "1", "L", false)
	pdfGenerator.SetY(currentPosY)
	pdfGenerator.MultiCell(0, 9, fmt.Sprintf("Created at: %v", time.Now().Format(viper.GetString("datetime_format"))), "1", "R", false)
	pdfGenerator.Ln(4)

	//issues
	for _, issue := range issues {
		// set issue font
		var lineHt float64 = 9
		pdfGenerator.SetFont("Arial", "", lineHt)
		pdfGenerator.SetTextColor(0, 0, 0)

		// parse issue template
		issueText := renderIssueToHTML(issue)
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
	err := pdfGenerator.OutputFileAndClose(fileName + ".pdf")

	if err != nil {
		errString := fmt.Sprintf("Erro while saving %s.pdf: %v", fileName, err)
		return errors.New(errString)
	}

	return nil
}

type fieldFunc func(jira.Issue) string

func renderIssueToHTML(issue jira.Issue) string {

	fieldMap := map[string]fieldFunc{
		"Key":     func(issue jira.Issue) string { return issue.Key },
		"Id":      func(issue jira.Issue) string { return issue.ID },
		"Summary": func(issue jira.Issue) string { return issue.Fields.Summary },
		"Assignee": func(issue jira.Issue) string {
			if issue.Fields.Assignee != nil {
				return issue.Fields.Assignee.Name
			}
			return ""
		},
		"Status": func(issue jira.Issue) string {
			if issue.Fields.Status != nil {
				return issue.Fields.Status.Name
			}
			return ""
		},
		"Description": func(issue jira.Issue) string { return issue.Fields.Description },
		"Created": func(issue jira.Issue) string {
			dateTime, err := time.Parse(viper.GetString("api_datetime_format"), issue.Fields.Created)
			if err != nil {
				log.Printf("Error on parse field \"created\"! %v\n", err)
				return ""
			}
			return dateTime.Format(viper.GetString("datetime_format"))
		},
		"Comment": func(issue jira.Issue) string {
			if issue.Fields.Comments == nil {
				return ""
			}
			var commentBuffer bytes.Buffer
			comments := issue.Fields.Comments.Comments
			if len(comments) > 0 {
				commentBuffer.WriteString("<br />")
			}
			for _, comment := range comments {
				c := fmt.Sprintf("<u>%s</u> - %s<br />", comment.Author.Name, comment.Body)
				commentBuffer.WriteString(c)
			}
			return commentBuffer.String()
		},
	}

	var issueText bytes.Buffer
	issueFields := viper.GetStringSlice("jira_issue_fields")

	for _, issueField := range issueFields {
		valueFunc, ok := fieldMap[issueField]
		if ok {
			issueText.WriteString(fmt.Sprintf("<b>%s:</b> %s<br />", issueField, valueFunc(issue)))
		}
	}

	return issueText.String()
}
