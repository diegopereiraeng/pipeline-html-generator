package main

import (
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"
)

const htmlTemplate = `
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            background-color: #f0f0f0;
        }
        .header {
            background-color: #0B5ED7; 
            color: white; 
            padding: 20px; 
            text-align: center; 
            font-size: 24px;
            border-bottom: 2px solid #fff;
        } 
        .section {
            padding: 20px;
            background-color: #fff;
            margin: 10px 20px;
            border-radius: 8px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
        }
        .green {
            background-color: rgba(40, 167, 69, 0.1);
        }
        .red {
            background-color: rgba(203, 36, 49, 0.1);
        }
        .orange {
            background-color: rgba(227, 98, 9, 0.1);
        }
        table {
            width: 100%; 
            border-collapse: collapse; 
            margin-bottom: 20px;
        }
        th, td {
            border: 1px solid #ccc; 
            padding: 8px; 
            text-align: left; 
        }
        th {
            background-color: #f8f8f8;
        }
    </style>
</head>
<body>
    <div class="header">
        Commit Insights Report
    </div>
    <div class="section">
        <strong>Repository Name:</strong> {{.RepoName}}<br>
        <strong>Branch Name:</strong> {{.BranchName}}<br>
        <strong>Trigger Type:</strong> {{.TriggerType}}<br>
    </div>
    <div class="section">
        <strong>Committers:</strong> {{.Committers}}
    </div>
    <div class="section">
        <strong>Participants:</strong> {{.Participants}}
    </div>
    <div class="section">
        <strong>File Changes:</strong>
        <table>
            <tr>
                <th>Committer/Reviewer</th>
                <th>Status</th>
                <th>File Name</th>
                <th>Commit Hash</th>
                <th>Title</th>
                <th>Date</th>
            </tr>
            {{range .FileChanges}}
            <tr class="{{.StatusClass}}">
                <td>{{.Committer}}{{if .Reviewer}} / {{.Reviewer}}{{end}}</td>
                <td>{{.Status}}</td>
                <td>{{.FileName}}</td>
                <td>{{.CommitHash}}</td>
                <td>{{.Title}}</td>
                <td>{{.Time}}</td>
            </tr>
            {{end}}
        </table>
    </div>
</body>
</html>
`

type reportData struct {
	RepoName     string
	BranchName   string
	TriggerType  string
	Committers   string
	Participants string
	FileChanges  []struct {
		FileName    string
		Status      string
		StatusClass string
		Committer   string
		Reviewer    string
		CommitHash  string
		Title       string
		Time        string
	}
}

func GenerateReport(repoName string, branchName string, serviceName string, committers []string, participants []string, fileChanges []struct {
	FileName   template.HTML
	Status     string
	Committer  string
	Reviewer   string
	CommitHash string
	Title      string
	Time       string
}) (string, error) {
	var committersStr, participantsStr string
	if len(committers) > 0 {
		committersStr = strings.Join(committers, ", ")
	}
	if len(participants) > 0 {
		participantsStr = strings.Join(participants, ", ")
	}

	var fileChangesData []struct {
		FileName    template.HTML
		Status      string
		StatusClass string
		Committer   string
		Reviewer    string
		CommitHash  string
		Title       string
		Time        string
	}
	for _, change := range fileChanges {
		var statusText, statusClass string
		switch change.Status {
		case "A":
			statusText = "Added"
			statusClass = "green"
		case "M":
			statusText = "Modified"
			statusClass = "orange"
		case "D":
			statusText = "Deleted"
			statusClass = "red"
		default:
			statusText = "Unknown"
		}
		//convert epoch 1694299436 format to human readable format
		timeInt, err := strconv.Atoi(change.Time)
		if err != nil {
			return "", err
		}
		date := time.Unix(int64(timeInt), 0).Format("2006-01-02 15:04:05")
		_ = date // fix "declared and not used" error
		fileChangesData = append(fileChangesData, struct {
			FileName    template.HTML
			Status      string
			StatusClass string
			Committer   string
			Reviewer    string
			CommitHash  string
			Title       string
			Time        string
		}{
			FileName:    change.FileName,
			Status:      statusText,
			StatusClass: statusClass,
			Committer:   change.Committer,
			Reviewer:    change.Reviewer,
			CommitHash:  change.CommitHash,
			Title:       change.Title,
			Time:        date,
		})
	}

	sort.Slice(fileChangesData, func(i, j int) bool {
		time1, _ := time.Parse("2006-01-02 15:04:05", fileChangesData[i].Time)
		time2, _ := time.Parse("2006-01-02 15:04:05", fileChangesData[j].Time)
		return time1.Before(time2)
	})

	data := reportData{
		RepoName:     repoName,
		BranchName:   branchName,
		Committers:   committersStr,
		Participants: participantsStr,
		FileChanges: func() []struct {
			FileName    string
			Status      string
			StatusClass string
			Committer   string
			Reviewer    string
			CommitHash  string
			Title       string
			Time        string
		} {
			var changes []struct {
				FileName    string
				Status      string
				StatusClass string
				Committer   string
				Reviewer    string
				CommitHash  string
				Title       string
				Time        string
			}
			for _, change := range fileChangesData {
				changes = append(changes, struct {
					FileName    string
					Status      string
					StatusClass string
					Committer   string
					Reviewer    string
					CommitHash  string
					Title       string
					Time        string
				}{
					FileName:    string(change.FileName),
					Status:      change.Status,
					StatusClass: change.StatusClass,
					Committer:   change.Committer,
					Reviewer:    change.Reviewer,
					CommitHash:  change.CommitHash, // Added this line
					Title:       change.Title,      // Added this line
					Time:        change.Time,       // Added this line
				})
			}
			return changes
		}(),
	}

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var report strings.Builder
	if err := tmpl.Execute(&report, data); err != nil {
		return "", err
	}

	return report.String(), nil
}
