package main

import (
	"encoding/json"
)

// Define structs to parse the JSON payload
type Actor struct {
	DisplayName string `json:"display_name"`
	Nickname    string `json:"nickname"`
	// Add other fields as needed
}

type Repository struct {
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
	// Add other fields as needed
}

type Push struct {
	Changes []Change `json:"changes"`
}

type Change struct {
	Old ChangeDetail `json:"old"`
	New ChangeDetail `json:"new"`
}

type ChangeDetail struct {
	Name   string `json:"name"`
	Target Target `json:"target"`
}

type Target struct {
	Hash string `json:"hash"`
	// Add other fields as needed
}

type Payload struct {
	Actor      Actor      `json:"actor"`
	Repository Repository `json:"repository"`
	Push       Push       `json:"push"`
}

// Function to parse the JSON payload and return the necessary details
func ParsePayload(jsonData []byte) (string, string, string, string, bool, error) {
	var payload Payload
	err := json.Unmarshal(jsonData, &payload)
	if err != nil {
		return "", "", "", "", false, err
	}
	oldHash := payload.Push.Changes[0].Old.Target.Hash
	newHash := payload.Push.Changes[0].New.Target.Hash
	branchName := payload.Push.Changes[0].New.Name
	repoName := payload.Repository.Name
	isPrivate := payload.Repository.IsPrivate

	return oldHash, newHash, branchName, repoName, isPrivate, nil
}
