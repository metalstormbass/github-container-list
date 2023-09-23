package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

func main() {
	username := os.Args[1]
	token := os.Getenv("GITHUB_TOKEN")
	ref := os.Args[2]
	recursionOption := os.Args[3]

	client := createGitHubClient(token)

	repos, err := getRepositories(client, username)
	if err != nil {
		log.Fatalf("Error getting repositories: %v", err)
	}

	for _, repo := range repos {

		getDockerfileContent(client, repo.GetName(), username, recursionOption, ref)

	}

}

func createGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func getRepositories(client *github.Client, username string) ([]*github.Repository, error) {
	ctx := context.Background()
	opt := &github.RepositoryListOptions{
		Type: "all",
	}

	var allRepos []*github.Repository
	page := 1

	for {
		repos, _, err := client.Repositories.List(ctx, username, opt)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)

		if len(repos) == 0 {
			break
		}

		page++
		opt.Page = page
	}

	return allRepos, nil
}

func getDockerfileContent(client *github.Client, repoFullName string, username string, recursionOption string, ref string) {
	ctx := context.Background()

	path := ""
	DockerfileNames := getDockerfileName(client, repoFullName, username, path, recursionOption, ref)

	if len(DockerfileNames) != 0 {

		for _, DockerfileName := range DockerfileNames {

			fileContent, _, _, err := client.Repositories.GetContents(ctx, username, repoFullName, DockerfileName, nil)
			if err != nil {

			}

			decodedContent, err := fileContent.GetContent()
			if err != nil {

			}

			container := findFROMLine(decodedContent)

			for containerName := range container {

				fmt.Println(container[containerName])
			}
		}
	}
	return
}

func getDockerfileName(client *github.Client, repoFullName string, username string, path string, recursionOption string, ref string) []string {
	ctx := context.Background()

	// Define the regular expression pattern to match Dockerfile variations
	pattern := `^(Dockerfile|.*\.Dockerfile|Dockerfile\..*)$`

	// Compile the regex pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		// Handle error
		return nil
	}

	// Recursively fetch the contents of the specified directory
	_, rootContents, _, err := client.Repositories.GetContents(ctx, username, repoFullName, path, &github.RepositoryContentGetOptions{Ref: ref})
	if err != nil {
		// Handle error
		return nil
	}

	// Initialize a slice to hold matching Dockerfile names
	var DockerfileNames []string

	// Iterate through the directory contents
	for _, content := range rootContents {
		if *content.Type == "file" {
			fileName := string(*content.Name)
			if regex.MatchString(fileName) {
				// Construct the full path for the matching Dockerfile
				filePath := fileName
				if path != "" {
					filePath = path + "/" + fileName
				}
				DockerfileNames = append(DockerfileNames, filePath)
			}
		} else if *content.Type == "dir" && recursionOption == "true" {
			// Recursively call the function for subdirectories
			subDirPath := path + "/" + string(*content.Name)
			subDirDockerfiles := getDockerfileName(client, repoFullName, username, subDirPath, recursionOption, ref)
			DockerfileNames = append(DockerfileNames, subDirDockerfiles...)
		}
	}

	return DockerfileNames
}

func findFROMLine(content string) []string {
	lines := strings.Split(content, "\n")
	var FROMline []string

	// Regular expression to match "FROM" lines and capture the container name
	regex := regexp.MustCompile(`\b([a-zA-Z0-9\-._]+:[a-zA-Z0-9\-._]+)\b`)

	for _, line := range lines {

		if strings.HasPrefix(strings.TrimSpace(line), "FROM ") {
			// Find matches in the line using the regular expression
			matches := regex.FindStringSubmatch(line)

			if len(matches) > 1 {
				// Extract the container name (group 1 in the regex match)
				containerName := matches[1]
				FROMline = append(FROMline, containerName)
			}
		}
	}

	return FROMline
}
