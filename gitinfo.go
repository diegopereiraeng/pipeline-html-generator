package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type CommitInfo struct {
	Hash           string
	Name           string
	Email          string
	Username       string
	AuthorTime     string
	CommitterName  string
	CommitterEmail string
	RefNames       string
	Title          string
	Body           string
	ParentHashes   string
	Changes        []FileChangeInfo
}

type FileChangeInfo struct {
	FileName string
	Status   string
}

type FileInfo struct {
	FileName      string
	CommitDetails []CommitInfo
}

func GetCommitInfo(olderCommitHash string, newerCommitHash string) ([]FileInfo, error) {
	fmt.Println("| \033[1;36mGetting commit info...\033[0m")
	//
	cmd := exec.Command("git", "log", "--pretty=format:%H;%an;%ae;%aN;%at;%cN;%cE;%d;%s;%b;%p", "--name-status", olderCommitHash+".."+newerCommitHash)
	fmt.Println("| \033[1;36mCommand:\033[0m " + cmd.String())

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// fmt.Println("|---------------------------------------------")
	// fmt.Println(out.String())
	// fmt.Println("|---------------------------------------------")

	fileCommitMap := make(map[string][]CommitInfo)
	lines := strings.Split(out.String(), "\n")
	var currentCommitInfo CommitInfo

	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.Contains(line, ";") {
			parts := strings.Split(line, ";")
			if len(parts) >= 11 {
				username := parts[3]
				emailRegex := regexp.MustCompile(`\+(\w+)@users\.noreply\.github\.com`)
				emailMatch := emailRegex.FindStringSubmatch(parts[2])
				if emailMatch != nil {
					username = emailMatch[1]
				}
				currentCommitInfo = CommitInfo{
					Hash:           parts[0],
					Name:           parts[1],
					Email:          parts[2],
					Username:       username,
					AuthorTime:     parts[4],
					CommitterName:  parts[5],
					CommitterEmail: parts[6],
					RefNames:       parts[7],
					Title:          parts[8],
					Body:           parts[9],
					ParentHashes:   parts[10],
					Changes:        []FileChangeInfo{},
				}
			}
		} else {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				changeInfo := FileChangeInfo{
					FileName: parts[1],
					Status:   parts[0],
				}
				currentCommitInfo.Changes = append(currentCommitInfo.Changes, changeInfo)
			}
		}
		if len(currentCommitInfo.Changes) > 0 {
			fileCommitMap[currentCommitInfo.Changes[0].FileName] = append(fileCommitMap[currentCommitInfo.Changes[0].FileName], currentCommitInfo)
			currentCommitInfo.Changes = currentCommitInfo.Changes[1:]
		}
	}

	var result []FileInfo
	for file, commits := range fileCommitMap {
		// fmt.Println("|---------------------------------------------")
		// fmt.Printf("| Commits: %v\n", len(commits))
		result = append(result, FileInfo{
			FileName:      file,
			CommitDetails: commits,
		})
		// fmt.Printf("| File: %s\n", file)
		// fmt.Println("|---------------------------------------------")
		// for _, commit := range commits {
		// 	fmt.Printf("| Commit: %s\n", commit.Hash)
		// 	fmt.Printf("| Author Name: %s\n", commit.Name)
		// 	fmt.Printf("| Author Email: %s\n", commit.Email)
		// 	fmt.Printf("| Username: %s\n", commit.Username)
		// 	fmt.Printf("| Author Time: %s\n", commit.AuthorTime)
		// 	fmt.Printf("| Committer Name: %s\n", commit.CommitterName)
		// 	fmt.Printf("| Committer Email: %s\n", commit.CommitterEmail)
		// 	fmt.Printf("| Ref Names: %s\n", commit.RefNames)
		// 	fmt.Printf("| Message Title: %s\n", commit.Title)
		// 	fmt.Printf("| Message Body: %s\n", commit.Body)
		// 	fmt.Printf("| Parent Hashes: %s\n", commit.ParentHashes)
		// 	fmt.Println("|---------------------------------------------")
		// 	for _, change := range commit.Changes {
		// 		fmt.Printf("| File: %s | Status: %s\n", change.FileName, change.Status)
		// 	}
		// 	fmt.Println("|---------------------------------------------")
		// }
	}

	return result, nil
}
