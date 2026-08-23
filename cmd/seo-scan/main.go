package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/zhaojunlucky/mkdocs-cms/core/seo"
)

func main() {
	root := flag.String("path", ".", "repository path to scan")
	allowMissing := flag.Bool("allow-missing", false, "exit successfully when mkdocs config is missing")
	failOn := flag.String("fail-on", "error", "minimum severity to fail on: error, warning, info, none")
	jsonOut := flag.Bool("json", false, "print JSON report")
	flag.Parse()

	report, err := seo.Scan(*root)
	if err != nil {
		if *allowMissing && err.Error() == "mkdocs config not found" {
			fmt.Println("SEO scan skipped: mkdocs config not found")
			return
		}
		fmt.Fprintf(os.Stderr, "SEO scan failed: %v\n", err)
		os.Exit(1)
	}

	if *jsonOut {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(report)
	} else {
		printReport(report)
	}

	if shouldFail(report.Issues, *failOn) {
		os.Exit(1)
	}
}

func printReport(report *seo.Report) {
	fmt.Printf("SEO scan: %d pages, %d issues\n", len(report.Pages), len(report.Issues))
	for _, issue := range report.Issues {
		if issue.Path == "" {
			fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.Type, issue.Message)
		} else {
			fmt.Printf("[%s] %s %s: %s\n", issue.Severity, issue.Type, issue.Path, issue.Message)
		}
	}
}

func shouldFail(issues []seo.Issue, failOn string) bool {
	threshold := severityRank(failOn)
	if threshold == 0 {
		return false
	}
	for _, issue := range issues {
		if severityRank(string(issue.Severity)) >= threshold {
			return true
		}
	}
	return false
}

func severityRank(severity string) int {
	switch severity {
	case "info":
		return 1
	case "warning":
		return 2
	case "error":
		return 3
	default:
		return 0
	}
}
