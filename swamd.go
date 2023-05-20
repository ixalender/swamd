package swamd

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func ParseGoFile(filePath, outputFileName string) {
	inFile, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", filePath, err)
		return
	}
	content, err := io.ReadAll(inFile)
	if err != nil {
		fmt.Printf("Error reading file content %s: %s\n", filePath, err)
		return
	}

	fset := token.NewFileSet()
	fileAst, err := parser.ParseFile(fset, "", string(content), parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file %s: %s\n", filePath, err)
		return
	}

	var annotatedComments []SwagAnnotationComment
	for _, commentGroup := range fileAst.Comments {
		for _, comment := range commentGroup.List {
			if hasSwaggerAnnotation(comment.Text) {
				cmt := extractAnnotationComment(comment.Text)
				annotatedComments = append(annotatedComments, *cmt)
			}
		}
	}

	outputFile, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Error opening output file %s: %s\n", outputFileName, err)
		return
	}
	defer outputFile.Close()

	if len(annotatedComments) > 0 {
		mdSpec := NewMarkdownAPISpec(annotatedComments)
		_, err = outputFile.WriteString(mdSpec.String() + "\n")
		if err != nil {
			fmt.Printf("Error writing to output file %s: %s\n", outputFileName, err)
			return
		}
	}

	fmt.Printf("%s - successfully processed.\n", filePath)
}

func hasSwaggerAnnotation(comment string) bool {
	annotationRegex := `@(Summary|Description|Tags|Accept|Produce|Param|Success|Router)`

	match, err := regexp.MatchString(annotationRegex, comment)
	if err != nil {
		fmt.Printf("Error detecting annotation: %s\n", err)
		return false
	}

	return match
}

func extractAnnotationComment(comment string) *SwagAnnotationComment {
	textRegex := `@(Summary|Description|Tags|Accept|Produce|Param|Success|Router)[[:space:]]+(.*)`

	r := regexp.MustCompile(textRegex)
	match := r.FindStringSubmatch(comment)
	if len(match) > 2 {
		return &SwagAnnotationComment{
			CommentType: match[1],
			CommentText: match[2],
		}
	}

	return nil
}

type SwagAnnotationComment struct {
	CommentType string
	CommentText string
}

type MarkdownAPISpec struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Accept      []string
	Produce     []string
	Params      []MarkdownParam
	Responses   []MarkdownResponse
}

type MarkdownParam struct {
	Name        string
	In          string
	Required    bool
	Type        string
	Description string
}

type MarkdownResponse struct {
	Code        int
	ParamType   string
	DataType    string
	Description string
}

func (m *MarkdownAPISpec) String() string {
	var markdown string
	markdown += fmt.Sprintf("|%s|%s|\n", strings.ToUpper(m.Method), m.Path)
	markdown += "| --- | --- |\n"
	if m.Summary != "" {
		markdown += fmt.Sprintf("|summary|%s|\n", m.Summary)
	}
	if m.Description != "" {
		markdown += fmt.Sprintf("|description|%s|\n", m.Description)
	}
	if len(m.Tags) > 0 {
		markdown += fmt.Sprintf("|Tags|%s|\n", strings.Join(m.Tags, ", "))
	}
	if len(m.Accept) > 0 {
		markdown += fmt.Sprintf("|Accept|%s|\n", strings.Join(m.Accept, ", "))
	}
	if len(m.Produce) > 0 {
		markdown += fmt.Sprintf("|Produce|%s|\n", strings.Join(m.Produce, ", "))
	}
	if len(m.Params) > 0 {
		markdown += "|Params"
		for i, param := range m.Params {
			mandatory := "optional"
			if param.Required {
				mandatory = "required"
			}
			borders := "|"
			if len(m.Params) > 1 && i > 0 {
				borders = "||"
			}
			markdown += fmt.Sprintf("%s**%s** %s {%s} %v – %s|\n",
				borders, param.Name, param.In, param.Type, mandatory, param.Description)
		}
	}
	if len(m.Responses) > 0 {
		markdown += "|Responses"
		for i, resp := range m.Responses {
			borders := "|"
			if len(m.Params) > 1 && i > 0 {
				borders = "||"
			}
			markdown += fmt.Sprintf("%s**%d** %s %s %s|\n",
				borders, resp.Code, resp.ParamType, resp.DataType, resp.Description)
		}
	}
	return markdown
}

func NewMarkdownAPISpec(swagAnnotationComments []SwagAnnotationComment) *MarkdownAPISpec {
	var markdownAPISpec MarkdownAPISpec

	for _, swagAnnot := range swagAnnotationComments {
		switch swagAnnot.CommentType {
		case "Summary":
			markdownAPISpec.Summary = swagAnnot.CommentText
		case "Description":
			markdownAPISpec.Description = swagAnnot.CommentText
		case "Tags":
			markdownAPISpec.Tags = append(markdownAPISpec.Tags, swagAnnot.CommentText)
		case "Accept":
			markdownAPISpec.Accept = append(markdownAPISpec.Accept, swagAnnot.CommentText)
		case "Produce":
			markdownAPISpec.Produce = append(markdownAPISpec.Produce, swagAnnot.CommentText)
		case "Param":
			p := NewMarkdownParam(swagAnnot.CommentText)
			if p != nil {
				markdownAPISpec.Params = append(markdownAPISpec.Params, *p)
			}
		case "Success", "Failure":
			s := NewMarkdownResponse(swagAnnot.CommentText)
			if s != nil {
				markdownAPISpec.Responses = append(markdownAPISpec.Responses, *s)
			}
		case "Router":
			markdownAPISpec.Path, markdownAPISpec.Method = parseRouter(swagAnnot.CommentText)
		}
	}

	return &markdownAPISpec
}

func NewMarkdownParam(commentText string) *MarkdownParam {
	paramRegex := `([[:alnum:]]+)[[:space:]]+([[:alnum:]]+)[[:space:]]+([[:alnum:]]+)[[:space:]]+(true|false)[[:space:]]+(.*)`
	r := regexp.MustCompile(paramRegex)
	match := r.FindStringSubmatch(commentText)

	if len(match) > 5 {
		return &MarkdownParam{
			Name:        match[1],
			In:          match[2],
			Type:        match[3],
			Required:    match[4] == "true",
			Description: match[5],
		}
	}

	return nil
}

func NewMarkdownResponse(commentText string) *MarkdownResponse {
	respRegex := `^(\d{3})\s*(\{[^\}]*\})?\s*(.*)$`
	r := regexp.MustCompile(respRegex)
	match := r.FindStringSubmatch(commentText)

	if len(match) >= 2 {
		code, err := strconv.Atoi(match[1])
		if err != nil {
			fmt.Printf("Error parsing response code: %s\n", err)
			return nil
		}
		respType := ""
		if len(match) >= 3 {
			respType = match[2]
		}
		respDataType := ""
		if len(match) >= 4 {
			respDataType = match[3]
		}
		description := ""
		if len(match) >= 5 {
			description = match[4]
		}
		return &MarkdownResponse{
			Code:        code,
			ParamType:   respType,
			DataType:    respDataType,
			Description: description,
		}
	}

	return nil
}

func parseRouter(commentText string) (string, string) {
	parts := strings.SplitN(commentText, " ", 2)
	path := strings.TrimSpace(parts[0])
	method := strings.TrimSpace(parts[1])

	return path, method
}
