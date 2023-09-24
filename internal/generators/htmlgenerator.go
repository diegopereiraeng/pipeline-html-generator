// generators/htmlgenerator.go
package htmlgenerator

import (
	"fmt"
	"html/template"
	"log"
	"pipeline-html-generator/internal/models"
	"sort"
	"strings"
	"time"
)

// GenerateDashboardHTML generates HTML for the dashboard based on a JSON structure.
func GenerateDashboardHTML(pipeline models.Pipeline) (string, error) {
	fmt.Println("|---------------------------------------------")
	fmt.Println("| \033[1;36mGenerating dashboard...\033[0m")
	fmt.Println("|---------------------------------------------")

	// fmt.Println("Pipeline: ", pipeline)

	const customDateFormat = "2006-01-02 15:04:05 -0700 MST"
	var err error
	startedTime, err := time.Parse(customDateFormat, pipeline.StartedTime)

	if err != nil {
		log.Println("Error parsing time: ", err)
	}
	pipeline.StartedTime = startedTime.Format("Jan 02 15:04:05 MST")

	sort.Slice(pipeline.Stages, func(i, j int) bool {
		startTimeI, err := time.Parse(customDateFormat, pipeline.Stages[i].StartTs)
		if err != nil && pipeline.Stages[i].StartTs != "" {
			log.Println("Error parsing time: ", err)
		}
		startTimeJ, err := time.Parse(customDateFormat, pipeline.Stages[j].StartTs)
		if err != nil && pipeline.Stages[j].StartTs != "" {
			log.Println("Error parsing time: ", err)
		}
		return startTimeI.Before(startTimeJ)
	})

	for i := range pipeline.Stages {
		sort.Slice(pipeline.Stages[i].Steps, func(m, n int) bool {
			startTimeM, errM := time.Parse(customDateFormat, pipeline.Stages[i].Steps[m].StartTs)
			startTimeN, errN := time.Parse(customDateFormat, pipeline.Stages[i].Steps[n].StartTs)

			if errM != nil && pipeline.Stages[i].Steps[m].Status != "Skipped" {
				log.Println("Status: ", pipeline.Stages[i].Steps[m].Status)
				log.Println("Error parsing time: ", errM)
			}
			if errN != nil && pipeline.Stages[i].Steps[n].Status != "Skipped" {
				log.Println("Status: ", pipeline.Stages[i].Steps[n].Status)
				log.Println("Error parsing time: ", errN)
			}

			if pipeline.Stages[i].Steps[m].Status == "Skipped" {
				return false
			}
			if pipeline.Stages[i].Steps[n].Status == "Skipped" {
				return true
			}

			return startTimeM.Before(startTimeN)
		})
	}

	for i := range pipeline.Stages {
		startTime, err := time.Parse(customDateFormat, pipeline.Stages[i].StartTs)
		if err != nil && pipeline.Stages[i].StartTs != "" {
			log.Println("Error parsing time: ", err)
		}
		endTime, err := time.Parse(customDateFormat, pipeline.Stages[i].EndTs)
		if err != nil && pipeline.Stages[i].EndTs != "" {
			log.Println("Error parsing time: ", err)
		}
		duration := endTime.Sub(startTime)

		if duration < time.Minute {
			pipeline.Stages[i].Duration = fmt.Sprintf("%.0f seconds", duration.Seconds())
		} else if duration < time.Hour {
			pipeline.Stages[i].Duration = fmt.Sprintf("%.0f minutes", duration.Minutes())
		} else if duration < time.Hour*24 {
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			pipeline.Stages[i].Duration = fmt.Sprintf("%d hours %d minutes", hours, minutes)
		} else if duration > time.Hour*24*365 {
			pipeline.Stages[i].Duration = fmt.Sprintln("None")
		} else {
			days := int(duration.Hours()) / 24
			hours := int(duration.Hours()) % 24
			pipeline.Stages[i].Duration = fmt.Sprintf("%d days %d hours", days, hours)

		}
		// convert start and end ts to shot date time and timezone , and year with 2 digits
		startTime = startTime.Local()
		pipeline.Stages[i].StartTs = startTime.Format("Jan 02 15:04:05 MST")
		endTime = endTime.Local()
		pipeline.Stages[i].EndTs = endTime.Format("Jan 02 15:04:05 MST")

		pipeline.Stages[i].Duration = duration.String()

		for j := range pipeline.Stages[i].Steps {
			startTime, err := time.Parse(customDateFormat, pipeline.Stages[i].Steps[j].StartTs)
			if err != nil && pipeline.Stages[i].Steps[j].StartTs != "" {
				log.Println("Error parsing time: ", err)
			}
			endTime, err := time.Parse(customDateFormat, pipeline.Stages[i].Steps[j].EndTs)
			if err != nil && pipeline.Stages[i].Steps[j].EndTs != "" {
				log.Println("Error parsing time: ", err)
			}

			pipeline.Stages[i].Steps[j].StartTs = startTime.Format("Jan02 15:04:05 MST")
			pipeline.Stages[i].Steps[j].EndTs = endTime.Format("Jan 02 15:04:05 MST")

			duration := endTime.Sub(startTime)

			if duration < time.Minute {
				pipeline.Stages[i].Steps[j].Duration = fmt.Sprintf("%.0f seconds", duration.Seconds())
			} else if duration < time.Hour {
				pipeline.Stages[i].Steps[j].Duration = fmt.Sprintf("%.0f minutes", duration.Minutes())
			} else if duration < time.Hour*24 {
				days := int(duration.Hours()) / 24
				pipeline.Stages[i].Steps[j].Duration = fmt.Sprintf("%d days", days)
			} else if duration > time.Hour*24*365 {
				pipeline.Stages[i].Steps[j].Duration = fmt.Sprintf("None")
			} else {
				days := int(duration.Hours()) / 24
				hours := int(duration.Hours()) % 24
				pipeline.Stages[i].Steps[j].Duration = fmt.Sprintf("%d days %d hours", days, hours)
			}
		}

	}

	const htmlTemplate = `
	<!DOCTYPE html>
	<html>
	<head>
		<style>
		body {
			font-family: Arial, sans-serif;
			margin: 0;
			background-color: #f0f0f0;
			display: flex;
			flex-direction: column;
			align-items: center;
		}
		.pipeline-container {
			background-color: #fff;
			border-radius: 10px;
			box-shadow: 0 0 10px rgba(0,0,0,0.1);
			overflow-x: auto;
			overflow-y: auto;
			max-width: 90%;
			max-height: 90vh;
			margin: 10px 20px;
			padding: 20px;
		}
		.pipeline-title {
			background-color: #00ABE3;
			color: white;
			padding: 20px;
			text-align: center;
			font-size: 20px; /* Adjusted font size */
			border-radius: 10px 10px 0 0;
			margin: -20px -20px 20px -20px;
		}
		.pipeline-info {
			font-size: 16px;
			color: #555;
			padding: 20px; /* Increased padding */
			background-color: #f8f8f8;
			border-bottom: 1px solid #ccc;
		}
		.stage-container {
			display: flex;
			flex-direction: row;
			align-items: flex-start;
			overflow-x: auto;
			overflow-y: auto;
			padding: 10px;
		}
		.stage {
			background-color: #f8f8f8;
			border: 1px solid #ccc;
			border-radius: 5px;
			padding: 10px;
			margin: 0 10px;
			box-shadow: 0 0 5px rgba(0,0,0,0.1);
			flex: 0 1 auto;
			width: 200px;  /* Increased width to accommodate the full width of steps */
			min-width: 200px; /* Adjusted min-width accordingly */
			max-width: 200px;
			font-size: 14px; /* Maintained font size */
		}
		.step-container {
			display: block;
			align-items: flex-start;
			width: 90%; /* Set to 100% to occupy full width of the stage */
			max-width: 100%; /* Set to 100% to prevent horizontal stretching */
			font-size: 12px; /* Consider increasing if text appears too small */
		}
		.step {
			display: block; /* Ensures it takes the full width available and stacks vertically */
			background-color: #fff;
			border: 1px solid #ccc;
			border-radius: 5px;
			padding: 8px;
			margin: 8px 0; /* Maintaining vertical margins */
			box-shadow: 0 0 5px rgba(0,0,0,0.1);
			width: 100%; /* Set to 100% to occupy the full width of the step container */
			max-width: 100%; /* Set to 100% to prevent horizontal stretching */
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
        .Success {
            background-color: rgba(76, 175, 80, 0.5); /* Green with Transparency */
        }
        .Failed {
            background-color: rgba(255, 87, 51, 0.5); /* Red with Transparency */
        }
        .Skipped {
            background-color: rgba(169, 169, 169, 0.5); /* Gray with Transparency */
        }
		.Aborted {
			background-color: rgba(255, 87, 51, 0.5); /* Red with Transparency */
		}
		</style>
	</head>
	<body>
	<div class="pipeline-container">
		<div class="pipeline-title">{{ .Name }} - Status: {{ .Status }}</div>
		<div class="pipeline-info">
			Started Time: {{ .StartedTime }}<br>
			Duration: {{ .Duration }}<br>
			Stage Count: {{ .StageCount }}<br>
			Step Count: {{ .StepCount }}<br>
			{{ if .Message }}Message: {{ .Message }}{{ end }}
		</div>
		<div class="stage-container">
			{{ range .Stages }}
			<div class="stage">
				<h4>{{ .Name }}</h4>
				<p>Duration: {{ .Duration }}</p>
				<div class="step-container">
					{{ range .Steps }}
					<div class="step {{ .Status }}">
						<h4>{{ .Name }}</h4>
						{{ if .Message }}<p>Message: {{ .Message }}</p>{{ end }}
						{{ if ne .Status "skipped" }}<p>Duration: {{ .Duration }}</p>{{ end }}
						{{ if eq .Status "failure" }}<p>Error: {{ .FailureInfo.Message }}</p><p>Failure Types: {{ range .FailureInfo.FailureTypeList }}{{ . }} {{ end }}</p>{{ end }}
					</div>
					{{ end }}
				</div>
			</div>
			{{ end }}
		</div>
	</div>
	</body>
	</html>
	`

	dataMap := map[string]interface{}{
		"Name":        pipeline.Name,
		"Status":      pipeline.Status,
		"StartedTime": pipeline.StartedTime,
		"Duration":    pipeline.Duration,
		"StageCount":  pipeline.StageCount,
		"StepCount":   pipeline.StepCount,
		"Message":     pipeline.Message,
		"Stages":      pipeline.Stages,
	}

	tmpl, err := template.New("dashboard").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var resultHTML strings.Builder
	err = tmpl.Execute(&resultHTML, dataMap)
	if err != nil {
		return "", err
	}

	return resultHTML.String(), nil
}
