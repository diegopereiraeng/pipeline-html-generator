// models/models.go
package models

// HTML GENERATOR

// Pipeline represents a pipeline with its stages and steps.
type Pipeline struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	StartedTime string  `json:"startedTime"`
	Duration    string  `json:"duration"`
	StageCount  int     `json:"stageCount"`
	StepCount   int     `json:"stepCount"`
	Message     string  `json:"message"`
	Stages      []Stage `json:"stages"`
}

// Stage represents a stage in a pipeline with its steps.
type Stage struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Module   string `json:"module"`
	StartTs  string `json:"startTs"`
	EndTs    string `json:"endTs"`
	Duration string `json:"duration"`
	Steps    []Step `json:"steps"`
}

// Step represents a step in a stage.
type Step struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	StartTs  string `json:"startTs"`
	EndTs    string `json:"endTs"`
	Duration string `json:"duration"`
}

// steps parsing
type PayloadSteps struct {
	Status string `json:"status"`
	Data   struct {
		ExecutionGraph struct {
			NodeMap map[string]Node `json:"nodeMap"`
		} `json:"executionGraph"`
	} `json:"data"`
}

type Node struct {
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	StartTs     int64  `json:"startTs"`
	EndTs       int64  `json:"endTs"`
	Status      string `json:"status"`
	StepType    string `json:"stepType"`
	FailureInfo struct {
		Message         string   `json:"message"`
		FailureTypeList []string `json:"failureTypeList"`
	} `json:"failureInfo"`
}

// PLUGIN CORE
