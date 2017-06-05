package pdf

import (
	"errors"
	"fmt"
	"log"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/viper"
)

const (
	lineHt float64 = 9
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
	fmt.Printf("Building %s.pdf issues:%d\n", fileName, len(issues))

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdfTR := pdf.UnicodeTranslatorFromDescriptor("")

	pdf.AddPage()

	// document title
	pdf.SetFont("Arial", "B", 16)
	pdf.SetFillColor(222, 222, 222)
	pdf.SetTextColor(0, 0, 0)
	pdf.MultiCell(0, 16, pdfTR(title+" Issues"), "1", "C", true)

	// document subtitle
	currentPosY := pdf.GetY()
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	pdf.MultiCell(0, 9, fmt.Sprintf("Total of issues: %v", len(issues)), "1", "L", false)
	pdf.SetY(currentPosY)
	pdf.MultiCell(0, 9, fmt.Sprintf("Created at: %v", time.Now().Format(viper.GetString("datetime_format"))), "1", "R", false)
	pdf.Ln(4)

	//issues
	for _, issue := range issues {
		// set issue font
		pdf.SetFont("Arial", "", lineHt)
		pdf.SetTextColor(0, 0, 0)

		// parse issue template
		renderIssueToHTML(issue, pdf)

		// draw issue separator
		pdf.SetDrawColor(195, 195, 195)
		pdf.Ln(lineHt)
		pdf.Ln(2)

		pageWidth, _ := pdf.GetPageSize()
		x, y := pdf.GetXY()
		marginL, marginR, _, _ := pdf.GetMargins()
		pdf.Line(x, y, x+pageWidth-marginR-marginL, y)

		pdf.Ln(2)
	}

	// save to output filename
	err := pdf.OutputFileAndClose(fileName + ".pdf")

	if err != nil {
		errString := fmt.Sprintf("Erro while saving %s.pdf: %v", fileName, err)
		return errors.New(errString)
	}

	return nil
}

type fieldFunc func(jira.Issue)

func renderIssueToHTML(issue jira.Issue, pdf *gofpdf.Fpdf) {
	maxChars := viper.GetInt("max_field_character_count")

	fieldMap := map[string]fieldFunc{
		"Key": func(issue jira.Issue) {
			pdf.Write(lineHt, "Key: ")
			pdf.Write(lineHt, issue.Key)
		},
		"Id": func(issue jira.Issue) {
			pdf.Write(lineHt, "Id: ")
			pdf.Write(lineHt, issue.ID)
		},
		"Summary": func(issue jira.Issue) {
			pdf.Write(lineHt, "Summary: ")
			pdf.Write(lineHt, issue.Fields.Summary)
		},
		"Assignee": func(issue jira.Issue) {
			if issue.Fields.Assignee != nil {
				pdf.Write(lineHt, "Assignee: ")
				pdf.Write(lineHt, issue.Fields.Assignee.Name)
			}
		},
		"Status": func(issue jira.Issue) {
			if issue.Fields.Status != nil {
				pdf.Write(lineHt, "Status: ")
				pdf.Write(lineHt, issue.Fields.Status.Name)
			}
		},
		"Description": func(issue jira.Issue) {
			pdf.Write(lineHt, "Description: ")

			if len(issue.Fields.Description) > maxChars {
				issue.Fields.Description = issue.Fields.Description[:maxChars]
			}
			pdf.Write(lineHt, issue.Fields.Description)
		},
		"Created": func(issue jira.Issue) {
			pdf.Write(lineHt, "Created: ")
			dateTime, err := time.Parse(viper.GetString("api_datetime_format"), issue.Fields.Created)
			if err != nil {
				log.Printf("Error on parse field \"created\"! %v\n", err)
				return
			}

			pdf.Write(lineHt, dateTime.Format(viper.GetString("datetime_format")))
		},
		"Comment": func(issue jira.Issue) {
			pdf.Write(lineHt, "Comments: ")

			if issue.Fields.Comments == nil {
				return
			}

			comments := issue.Fields.Comments.Comments
			if len(comments) > 0 {
				pdf.Ln(lineHt)
			}

			for _, comment := range comments {
				if len(comment.Body) > maxChars {
					comment.Body = comment.Body[:maxChars]
				}
				pdf.Write(lineHt, comment.Author.Name)
				pdf.Write(lineHt, " - ")
				pdf.Write(lineHt, comment.Body)
				pdf.Ln(lineHt)
			}
		},
	}

	issueFields := viper.GetStringSlice("jira_issue_fields")

	for _, issueField := range issueFields {
		valueFunc, ok := fieldMap[issueField]
		if ok {
			valueFunc(issue)
			pdf.Ln(lineHt)
		}
	}
}
