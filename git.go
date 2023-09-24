package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Committer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func git() {
	cmd := exec.Command("git", "log", "--pretty=format:%an <%ae>", "e7c79ef9dcaa60c41c88ea5417b977bffe0bdb9f..HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return
	}
	fmt.Println("| Getting commiters...")

	commiterMap := make(map[string]Committer)
	outputLines := strings.Split(out.String(), "\n")
	for _, line := range outputLines {
		if line != "" {
			parts := strings.SplitN(line, " <", 2)
			if len(parts) == 2 {
				email := strings.TrimSuffix(parts[1], ">")
				commiterMap[email] = Committer{Name: parts[0], Email: email}
			}
		}
	}

	commiters := make([]Committer, 0, len(commiterMap))
	for _, commiter := range commiterMap {
		commiters = append(commiters, commiter)
	}

	jsonOutput, err := json.Marshal(commiters)
	if err != nil {
		fmt.Println("Error marshalling output:", err)
		return
	}
	fmt.Println("|---------------------------------------------")
	fmt.Println(string(jsonOutput))
	fmt.Println("|---------------------------------------------")
}
