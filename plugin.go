package main

import (
	"io"
	"path/filepath"
	"pipeline-html-generator/internal/models"
	"strconv"

	htmlgenerator "pipeline-html-generator/internal/generators"

	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// define Config and Plugin structure like this

type (
	Config struct {
		AccID            string   `json:"accID"`
		OrgID            string   `json:"orgID"`
		ProjectID        string   `json:"projectID"`
		PipelineID       string   `json:"pipelineID"`
		StatusList       []string `json:"statusList"`
		RepoName         string   `json:"repoName"`
		Branch           string   `json:"branch"`
		ServiceName      string   `json:"serviceName"`
		HarnessSecret    string   `json:"harnessSecret"`
		PipeExecutionURL string   `json:"harnessPipeExecutionURL"`
	}

	Plugin struct {
		Config Config
	}
)

type Commit struct {
	Recast string `json:"__recast"`
	ID     string `json:"id"`
	// Include other fields if needed
}

type BranchInfo struct {
	Recast  string   `json:"__recast"`
	Commits []Commit `json:"commits"`
	// Include other fields if needed
}

type CIExecutionInfoDTO struct {
	Branch BranchInfo `json:"branch"`
	// Include other fields if needed
}

type Branch struct {
	Recast  string   `json:"__recast"`
	Commits []Commit `json:"commits"`
}

type CI struct {
	CIExecutionInfoDTO CIExecutionInfoDTO `json:"ciExecutionInfoDTO"`
	// Branch             BranchInfo         `json:"branch"`
	// Include other fields if needed
}

type ModuleInfo struct {
	CI CI `json:"ci"`
	// Include other fields if needed
}

type Content struct {
	ModuleInfo            ModuleInfo    `json:"moduleInfo"`
	LayoutNodeMap         LayoutNodeMap `json:"layoutNodeMap"`
	PlanExecutionId       string        `json:"planExecutionId"`
	Status                string        `json:"status"`
	Name                  string        `json:"name"`
	StartTs               int           `json:"startTs"`
	EndTs                 int           `json:"endTs"`
	SuccessfulStagesCount int           `json:"successfulStagesCount"`
	FailedStagesCount     int           `json:"failedStagesCount"`
	TotalStagesCount      int           `json:"totalStagesCount"`
	StartingNodeId        string        `json:"startingNodeId"`
	ExecutionTriggerInfo  struct {
		TriggerType string `json:"triggerType"`
		TriggeredBy struct {
			Identifier string `json:"identifier"`
			ExtraInfo  struct {
				Email string `json:"email"`
			}
		}
		IsRerun bool `json:"isRerun"`
	} `json:"executionTriggerInfo"`
	// Include other fields if needed
}

type Data struct {
	Content []Content `json:"content"`
	// Include other fields if needed
}

type Response struct {
	Data   Data   `json:"data"`
	Status string `json:"status"`
	// Include other fields if needed
}

type NodeInfo struct {
	NodeType       string     `json:"nodeType"`
	NodeGroup      string     `json:"nodeGroup"`
	NodeIdentifier string     `json:"nodeIdentifier"`
	Name           string     `json:"name"`
	NodeUuid       string     `json:"nodeUuid"`
	Status         string     `json:"status"`
	Module         string     `json:"module"`
	ModuleInfo     ModuleInfo `json:"moduleInfo"`
	StartTs        int        `json:"startTs"`
	EndTs          int        `json:"endTs"`
	FailureInfo    struct {
		Message string `json:"message"`
	} `json:"failureInfo"`
	EdgeLayoutList EdgeLayout `json:"edgeLayoutList"`
	// NodeExecutionId string `json:"nodeExecutionId"`
	// Include other fields if needed
}

type EdgeLayout struct {
	CurrentNodeChildren []string `json:"currentNodeChildren"`
	NextIds             []string `json:"nextIds"`
	// ... (other fields)
}

type LayoutNodeMap map[string]NodeInfo

var plugin Plugin

const lineBreak = "|---------------------------------------------"

func getExecutionDetails(accID string, orgID string, projectID string, pipelineID string, statusList []string, repoName string, branch string, serviceName string) (models.Pipeline, error) {

	url := "https://app.harness.io/pipeline/api/pipelines/execution/summary?page=0&size=1&accountIdentifier=" + accID + "&orgIdentifier=" + orgID + "&projectIdentifier=" + projectID + "&pipelineIdentifier=" + pipelineID + ""
	method := "POST"

	var statusListJson string
	statusListJsonBytes, err := json.Marshal(statusList)
	if err != nil {
		return models.Pipeline{}, err
	}
	statusListJson = string(statusListJsonBytes)

	payload := strings.NewReader(fmt.Sprintf(`{"status":%s,"moduleProperties":{"ci":{"branch":"%s","repoName":"%s"}},"filterType":"PipelineExecution"}`, statusListJson, branch, repoName))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println("Error finding last successful execution")
		fmt.Println("URL: ", url)
		fmt.Println("Payload: ", payload)
		fmt.Println("Error: ", err)
		return models.Pipeline{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", plugin.Config.HarnessSecret)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error finding last successful execution")
		fmt.Println("URL: ", url)
		fmt.Println("Payload: ", payload)
		fmt.Println("Error: ", err)
		fmt.Println("Status: ", req.Response.Status)
		return models.Pipeline{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error parsing response")
		fmt.Println("URL: ", url)
		fmt.Println("Payload: ", payload)
		fmt.Println("Response: ", body)
		fmt.Println("Error: ", err)
		fmt.Println("Status: ", req.Response.Status)
		return models.Pipeline{}, err
	}
	fmt.Println("URL: ", url)
	fmt.Println("Payload: ", payload)
	fmt.Println("Response: ", string(body))

	defer res.Body.Close()

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return models.Pipeline{}, errors.New("error parsing JSON response from Harness API Pipeline Executions")
	}

	if len(response.Data.Content) == 0 {
		return models.Pipeline{}, errors.New("no successful execution found")
	}

	// responseBytes, err := json.Marshal(response)
	// if err != nil {
	// 	return models.Pipeline{}, err
	// }
	// pipelineParsed, err := parsePipeline(responseBytes)
	// if err != nil {
	// 	return models.Pipeline{}, err
	// }

	// fmt.Println("Pipeline: ", pipelineParsed)
	// fmt.Println("Pipeline Content: ", response.Data.Content)

	var pipeline models.Pipeline
	for _, content := range response.Data.Content {
		fmt.Printf("| Found execution with status:\033[0m \033[1;32m%s\033[0m\n", content.Status)
		fmt.Println(lineBreak)
		fmt.Printf("| \033[1;36mPlan Execution ID:\033[0m \033[1;32m%s\033[0m\n", content.PlanExecutionId)
		fmt.Printf("| \033[1;36mPipeline Name:\033[0m \033[1;32m%s\033[0m\n", content.Name)
		fmt.Printf("| \033[1;36mPipe Status:\033[0m \033[1;32m%s\033[0m\n", content.Status)
		fmt.Println(lineBreak)

		if content.Status == "Running" {
			content.EndTs = int(time.Now().UnixNano() / int64(time.Millisecond))
		}

		pipeline = models.Pipeline{
			Name:        content.Name,
			Status:      content.Status,
			StartedTime: time.Unix(int64(content.StartTs/1000), 0).String(),
			Duration:    time.Unix(int64(content.EndTs/1000), 0).Sub(time.Unix(int64(content.StartTs/1000), 0)).String(),
			StageCount:  0,
			StepCount:   0,
			Message:     "",
			ExecutionId: content.PlanExecutionId,
		}

		// content := response.Data.Content[0]
		// foundSuccessfulExecution := false

		for _, nodeInfo := range content.LayoutNodeMap {
			fmt.Println(lineBreak)
			fmt.Printf("| \033[1;36mNode Type:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.NodeType)
			fmt.Printf("| \033[1;36mNode Group:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.NodeGroup)
			fmt.Printf("| \033[1;36mNode Identifier:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.NodeIdentifier)
			fmt.Printf("| \033[1;36mName:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.Name)
			fmt.Printf("| \033[1;36mNode UUID:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.NodeUuid)
			fmt.Printf("| \033[1;36mStatus:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.Status)
			fmt.Printf("| \033[1;36mModule:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.Module)
			fmt.Printf("| \033[1;36mStart TS:\033[0m \033[1;32m%s\033[0m\n", time.Unix(int64(nodeInfo.StartTs/1000), 0))
			fmt.Printf("| \033[1;36mEnd TS:\033[0m \033[1;32m%s\033[0m\n", time.Unix(int64(nodeInfo.EndTs/1000), 0))
			fmt.Printf("| \033[1;36mFailure Info:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.FailureInfo.Message)
			fmt.Printf("| \033[1;36mEdge Layout List:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.EdgeLayoutList)
			fmt.Println(lineBreak)
			stageNodeID := nodeInfo.NodeUuid
			urlSteps := "https://app.harness.io/gateway/pipeline/api/pipelines/execution/v2/" + content.PlanExecutionId + "?page=0&size=1&accountIdentifier=" + accID + "&orgIdentifier=" + orgID + "&projectIdentifier=" + projectID + "&pipelineIdentifier=" + pipelineID + "&stageNodeId=" + stageNodeID + ""
			methodSteps := "GET"
			clientSteps := &http.Client{}
			reqSteps, err := http.NewRequest(methodSteps, urlSteps, nil)

			if err != nil {
				return models.Pipeline{}, err
			}
			reqSteps.Header.Add("Content-Type", "application/json")
			reqSteps.Header.Add("x-api-key", plugin.Config.HarnessSecret)

			resSteps, err := clientSteps.Do(reqSteps)
			if err != nil {
				return models.Pipeline{}, err
			}

			bodySteps, err := io.ReadAll(resSteps.Body)

			if err != nil {
				return models.Pipeline{}, err
			}

			// fmt.Printf("| \033[1;36mResponse Body:\033[0m \033[1;32m%s\033[0m\n", string(bodySteps))
			defer resSteps.Body.Close()

			var payloadSteps models.PayloadSteps
			err = json.Unmarshal(bodySteps, &payloadSteps)
			if err != nil {
				return models.Pipeline{}, errors.New("error parsing JSON Stage Details response from Harness API Pipeline Executions")
			}
			// also check if number of step is less than 1 (payloadSteps.Data.ExecutionGraph.NodeMap[])
			if nodeInfo.Name != "" && nodeInfo.NodeType != "STEP_GROUP" && nodeInfo.NodeType != "NG_FORK" && nodeInfo.NodeType != "ROLLBACK_OPTIONAL_CHILD_CHAIN" && nodeInfo.Status != "NotStarted" && nodeInfo.Status != "Skipped" {

				var startTS string
				var endTS string
				var duration string

				if nodeInfo.Status == "Skipped" {
					startTS = ""
					endTS = ""
					duration = "0s"
				} else if nodeInfo.Status == "Running" || nodeInfo.Status == "AsyncWaiting" {
					nodeInfo.EndTs = int(time.Now().UnixNano() / int64(time.Millisecond))
					startTS = time.Unix(int64(nodeInfo.StartTs/1000), 0).String()
					endTS = time.Unix(int64(nodeInfo.EndTs/1000), 0).String()
					// use now as end time
					duration = time.Unix(int64(time.Now().UnixNano()/1000), 0).Sub(time.Unix(int64(nodeInfo.StartTs/1000), 0)).String()
				} else {
					startTS = time.Unix(int64(nodeInfo.StartTs/1000), 0).String()
					endTS = time.Unix(int64(nodeInfo.EndTs/1000), 0).String()
					duration = time.Unix(int64(nodeInfo.EndTs/1000), 0).Sub(time.Unix(int64(nodeInfo.StartTs/1000), 0)).String()
				}

				pipeline.Stages = append(pipeline.Stages, models.Stage{
					Name:     nodeInfo.Name,
					Status:   nodeInfo.Status,
					Module:   nodeInfo.Module,
					Steps:    []models.Step{},
					StartTs:  startTS,
					EndTs:    endTS,
					Duration: duration,
				})

				for _, node := range payloadSteps.Data.ExecutionGraph.NodeMap {
					fmt.Println(lineBreak)
					fmt.Printf("| \033[1;36mStep Name:\033[0m \033[1;32m%s\033[0m\n", node.Name)
					fmt.Printf("| \033[1;36mStep Identifier:\033[0m \033[1;32m%s\033[0m\n", node.Identifier)
					if node.Status != "Skipped" {
						fmt.Printf("| \033[1;36mStep Start TS:\033[0m \033[1;32m%s\033[0m\n", time.Unix(int64(node.StartTs/1000), 0))
						fmt.Printf("| \033[1;36mStep End TS:\033[0m \033[1;32m%s\033[0m\n", time.Unix(int64(node.EndTs/1000), 0))
						fmt.Printf("| \033[1;36mStep Duration:\033[0m \033[1;32m%s\033[0m\n", time.Unix(int64(node.EndTs/1000), 0).Sub(time.Unix(int64(node.StartTs/1000), 0)))
					}
					fmt.Printf("| \033[1;36mStep Status:\033[0m \033[1;32m%s\033[0m\n", node.Status)
					fmt.Printf("| \033[1;36mStep Type:\033[0m \033[1;32m%s\033[0m\n", node.StepType)
					if node.FailureInfo.Message != "" {
						fmt.Printf("| \033[1;36mStep Failure Info:\033[0m \033[1;32m%s\033[0m\n", node.FailureInfo.Message)
						fmt.Printf("| \033[1;36mStep Failure Type List:\033[0m \033[1;32m%s\033[0m\n", node.FailureInfo.FailureTypeList)
					}
					fmt.Println(lineBreak)
					if node.Identifier != "execution" && node.Name != "parallel" && node.Name != "liteEngineTask" && node.StepType != "STEP_GROUP" && node.StepType != "NG_FORK" && node.StepType != "ROLLBACK_OPTIONAL_CHILD_CHAIN" && node.StepType != "IntegrationStageStepPMS" && node.Status != "NotStarted" && node.Status != "Skipped" {
						var startTS string
						var endTS string
						var duration string
						var message string
						var status string
						if node.Status == "Skipped" {
							startTS = ""
							endTS = ""
							duration = "0s"
						} else {
							startTS = time.Unix(int64(node.StartTs/1000), 0).String()
							if node.Status == "Running" || node.Status == "AsyncWaiting" {
								endTS = time.Unix(0, time.Now().UnixNano()).String()
								node.EndTs = time.Now().UnixNano() / int64(time.Millisecond)
								duration = time.Unix(int64(time.Now().UnixNano()/1000), 0).Sub(time.Unix(int64(node.StartTs/1000), 0)).String()
							} else {
								endTS = time.Unix(int64(node.EndTs/1000), 0).String()
								duration = time.Unix(int64(node.EndTs/1000), 0).Sub(time.Unix(int64(node.StartTs/1000), 0)).String()
							}

						}

						if node.Status != "Success" && node.FailureInfo.Message != "" {
							message = node.FailureInfo.Message
							status = node.Status
						} else if node.Status == "Success" && node.FailureInfo.Message != "" {
							message = "Ignored Error"
							// status = "Success - Error Ignored"
						} else {
							message = node.FailureInfo.Message
							status = node.Status
						}

						pipeline.Stages[pipeline.StageCount].Steps = append(pipeline.Stages[pipeline.StageCount].Steps, models.Step{
							Name:        node.Name,
							Status:      status,
							Message:     message,
							StartTs:     startTS,
							EndTs:       endTS,
							Duration:    duration,
							FailureInfo: node.FailureInfo,
						})
						pipeline.StepCount++
					}
				}

				pipeline.StageCount++
			}
		}

		// fmt.Printf("| \033[1;36mStage ID:\033[0m \033[1;32m%s\033[0m\n", stageID)

		var commiters []string

		for _, commit := range content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits {
			commiters = append(commiters, commit.ID)
		}
		// fmt.Printf("Commits: %s\n", strings.Join(commiters, ", "))
		fmt.Printf("| \033[1;36mNumber of commits:\033[0m \033[1;32m%d\033[0m\n", len(content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits))
		fmt.Println(lineBreak)
		// fmt.Printf("Last Commit SHA: %s\n", content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits[len(content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits)-1].ID)

		if len(content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits) > 0 {
			// fmt.Printf("First Commit SHA: %s\n", content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits[0].ID)
		} else {
			fmt.Println("No commits found")
		}

		return pipeline, nil
	}

	return models.Pipeline{}, errors.New("no successful execution found")
}

func (p *Plugin) Exec() error {

	plugin = *p

	var accID string = p.Config.AccID
	var orgID string = p.Config.OrgID
	var projectID string = p.Config.ProjectID
	var pipelineID string = p.Config.PipelineID
	var statusList []string = p.Config.StatusList
	var repoName string = p.Config.RepoName
	var branch string = p.Config.Branch
	var serviceName string = p.Config.ServiceName

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36m Pipeline HTML Generator Plugin\033[0m")
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mDeveloped By:\033[0m \033[1;32mDiego Pereira\033[0m")
	fmt.Println(lineBreak)
	fmt.Printf("| \033[1;36mAccount ID:\033[0m \033[1;32m%s\033[0m\n", accID)
	fmt.Printf("| \033[1;36mOrg ID:\033[0m \033[1;32m%s\033[0m\n", orgID)
	fmt.Printf("| \033[1;36mProject ID:\033[0m \033[1;32m%s\033[0m\n", projectID)
	fmt.Printf("| \033[1;36mPipeline ID:\033[0m \033[1;32m%s\033[0m\n", pipelineID)
	fmt.Printf("| \033[1;36mStatus List:\033[0m \033[1;32m%s\033[0m\n", statusList)
	fmt.Println(lineBreak)
	if repoName == "" {
		fmt.Printf("| \033[1;36mService Name:\033[0m \033[1;32m%s\033[0m\n", serviceName)
	} else {
		fmt.Printf("| \033[1;36mRepo Name:\033[0m \033[1;32m%s\033[0m\n", repoName)
		fmt.Printf("| \033[1;36mBranch:\033[0m \033[1;32m%s\033[0m\n", branch)
	}

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mSearching for execution details...\033[0m")
	fmt.Println(lineBreak)

	var pipeline models.Pipeline

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mGetting last successful execution...\033[0m")
	fmt.Println(lineBreak)
	// Get the old and new commit hashes from the pipeline
	var err error
	pipeline, err = getExecutionDetails(accID, orgID, projectID, pipelineID, statusList, repoName, branch, serviceName)
	if err != nil {
		fmt.Println(" | \033[1;31mError getting execution details\033[0m")
		fmt.Println(" | \033[1;31mError: ", err, "\033[0m")
		// return err
	}
	fmt.Println(lineBreak)
	if pipeline.Status != "" {
		fmt.Println("| \033[1;32mLast successful execution found\033[0m")
		fmt.Println(lineBreak)
		fmt.Printf("| \033[1;36mPipeline Name:\033[0m \033[1;32m%s\033[0m\n", pipeline.Name)
		fmt.Printf("| \033[1;36mPipeline Status:\033[0m \033[1;32m%s\033[0m\n", pipeline.Status)
		fmt.Printf("| \033[1;36mPipeline Started Time:\033[0m \033[1;32m%s\033[0m\n", pipeline.StartedTime)
		fmt.Printf("| \033[1;36mPipeline Duration:\033[0m \033[1;32m%s\033[0m\n", pipeline.Duration)
		fmt.Printf("| \033[1;36mPipeline Stage Count:\033[0m \033[1;32m%d\033[0m\n", pipeline.StageCount)
		fmt.Printf("| \033[1;36mPipeline Step Count:\033[0m \033[1;32m%d\033[0m\n", pipeline.StepCount)
		fmt.Printf("| \033[1;36mPipeline Message:\033[0m \033[1;32m%s\033[0m\n", pipeline.Message)
		fmt.Println(lineBreak)
	} else {
		fmt.Println("| \033[1;31mLast successful execution not found\033[0m")
		fmt.Println(lineBreak)
		return errors.New("successful execution not found")
	}

	fmt.Println(lineBreak)
	// Create USer Execution Link
	executionLink := "https://app.harness.io/ng/account/" + accID + "/ci/orgs/" + orgID + "/projects/" + projectID + "/pipelines/" + pipelineID + "/deployments/" + pipeline.ExecutionId + "/pipeline"
	pipeline.ExecutionLink = executionLink
	dashHTML, err := htmlgenerator.GenerateDashboardHTML(pipeline)
	if err != nil {
		return err
	}

	//save to a html file
	err = os.WriteFile("pipeline.html", []byte(dashHTML), 0644)
	if err != nil {
		return err
	}

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mPipeline HTML Generator saved to pipeline.html\033[0m")
	fmt.Println(lineBreak)
	// save to env file
	vars := map[string]string{
		"PIPELINE_NAME":        pipeline.Name,
		"PIPELINE_STATUS":      pipeline.Status,
		"PIPELINE_STARTEDTIME": pipeline.StartedTime,
		"PIPELINE_DURATION":    pipeline.Duration,
		"PIPELINE_STAGECOUNT":  strconv.Itoa(pipeline.StageCount),
		"PIPELINE_STEPCOUNT":   strconv.Itoa(pipeline.StepCount),
		"PIPELINE_MESSAGE":     pipeline.Message,
		"HTML_REPORT":          strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(dashHTML, "\n", ""), "	", ""), "		", ""), "<!DOCTYPE html>", ""),
	}

	err = writeEnvFile(vars, os.Getenv("DRONE_OUTPUT"))
	if err != nil {
		// return err
		fmt.Println("| \033[33m[WARNING] - Error writing to .env file: ", err, "\033[0m")

	}

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mPipeline Env File saved to pipeline.env\033[0m")
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mDeveloped by: \033[0m \033[1;32mDiego Pereira\033[0m")
	fmt.Println("| \033[1;36mGithub: \033[0m \033[1;32mhttps://github.com/diegopereiraeng\033[0m")
	fmt.Println("| \033[1;36mLinkedIn: \033[0m \033[1;32mhttps://www.linkedin.com/in/diego-pereira-eng\033[0m")
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mPipeline HTML Generator Plugin Completed\033[0m")
	fmt.Println(lineBreak)

	return nil
}

func writeEnvFile(vars map[string]string, outputPath string) error {
	if outputPath == "" {
		return writeDefaultEnvFile(vars)
	}
	return writeSpecifiedEnvFile(vars, outputPath)
}

func writeDefaultEnvFile(vars map[string]string) error {
	outputPath := "PipelineHTMLGenerator.env"
	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer f.Close()

	for key, value := range vars {
		f.WriteString(fmt.Sprintf("export %s='%s'\n", key, value))
	}

	return nil
}

func writeSpecifiedEnvFile(vars map[string]string, outputPath string) error {
	if err := createDirIfNotExists(outputPath); err != nil {
		return err
	}

	if err := createFileIfNotExists(outputPath); err != nil {
		return err
	}

	return writeVarsUsingGodotenv(vars, outputPath)
}

func createDirIfNotExists(outputPath string) error {
	dir := filepath.Dir(outputPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Creating directory:", dir)
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func createFileIfNotExists(outputPath string) error {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		fmt.Println("| Creating env file for Harness:", outputPath)
		_, err := os.Create(outputPath)
		if err != nil {
			fmt.Println("| \033[33m[WARNING] - Error creating file:", err, "\033[0m")
		}
		return err
	}
	return nil
}

func writeVarsUsingGodotenv(vars map[string]string, outputPath string) error {
	err := godotenv.Write(vars, outputPath)
	if err != nil {
		fmt.Println("| \033[33m[WARNING] Error writing to .env file:", err, "\033[0m")
		return err
	}
	fmt.Println("| Successfully wrote to .env file")
	return nil
}

// func Parse models.PayloadSteps(jsonData []byte) (* models.PayloadSteps, error) {
func ParsePayloadSteps() (*models.PayloadSteps, error) {
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mParsing payload...\033[0m")
	fmt.Println(lineBreak)

	file, err := os.Open("pipegraph.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	// Read the file content
	data, err := os.ReadFile("pipegraph.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	var payloadSteps models.PayloadSteps
	//convert data to strings
	// fmt.Println("Data: ", string(data))
	// fmt.Println("Data: ", data)
	err = json.Unmarshal(data, &payloadSteps)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil, err
	}

	for _, node := range payloadSteps.Data.ExecutionGraph.NodeMap {
		fmt.Printf("Step: %s (%s)\n", node.Name, node.Identifier)
		fmt.Printf("Status: %s\n", node.Status)
		if node.Status == "Failed" {
			fmt.Printf("Failure Message: %s\n", node.FailureInfo.Message)
			fmt.Printf("Failure Types: %v\n", node.FailureInfo.FailureTypeList)
		}
		fmt.Printf("Start Time: %s\n", time.Unix(0, node.StartTs*int64(time.Millisecond)).UTC())
		fmt.Printf("End Time: %s\n", time.Unix(0, node.EndTs*int64(time.Millisecond)).UTC())
		fmt.Println("---")
	}

	return &payloadSteps, nil
}
