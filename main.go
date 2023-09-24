package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

func main() {
	// Check the number of command-line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: program_name username branch [recursionOption] [iteration]")
		os.Exit(1)
	}

	// Username or Org
	username := os.Args[1]

	// Branch
	ref := os.Args[2]
	if ref == "" {
		fmt.Println("Please specify a branch")
		os.Exit(1)
	}

	// Github Token (optional)
	token := os.Getenv("GITHUB_TOKEN")

	// Recursion Option (optional)
	var recursionOption string
	if len(os.Args) > 3 {
		recursionOption = os.Args[3]
		if recursionOption != "true" && recursionOption != "false" {
			fmt.Println("Enter true or false as values for recursion")
			os.Exit(1)
		}
	} else {
		recursionOption = "false"
	}

	// Iteration, for continuation after hitting rate limit (optional)
	var iterationStr string
	if len(os.Args) > 4 {
		iterationStr = os.Args[4]
	} else {
		iterationStr = "0"
	}

	// Convert Iteration to Integer
	iteration, err := strconv.Atoi(iterationStr)
	if err != nil {
		fmt.Println("Invalid iteration value:", err)
		os.Exit(1)
	}

	// Create Github Client
	client := createGitHubClient(token)

	// Get List of Repos
	repos, err := getRepositories(client, username)
	if err != nil {
		log.Println(err)
	}

	// Loop through repos to search for Dockerfiles
	for range repos {
		for {
			// Check to see if a repo contains any Dockerfiles
			hasDockerfile, err := hasDockerfiles(client, username, repos[iteration].GetName())
			if err != nil {
				handleRateLimit(err, repos[iteration].GetName(), iteration)
			}

			// Parse any found Dockerfiles
			if hasDockerfile {
				getDockerfileContent(client, repos[iteration].GetName(), username, recursionOption, ref)
			}
			iteration++
			if iteration >= len(repos) {
				log.Println("Complete")
				os.Exit(0)
			}
		}

	}

}

// Function to create Github Client
func createGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

// Function to get repos
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

// Function to check if repo has Dockefiles
func hasDockerfiles(client *github.Client, username, repo string) (bool, error) {
	ctx := context.Background()

	// Retrieve the repository's languages
	languages, _, err := client.Repositories.ListLanguages(ctx, username, repo)
	if err != nil {
		return false, err
	}

	// Check if "Dockerfile" language is present
	_, ok := languages["Dockerfile"]
	return ok, nil
}

// Function to parse Dockerfiles
func getDockerfileContent(client *github.Client, repoFullName string, username string, recursionOption string, ref string) {
	ctx := context.Background()

	path := ""

	// Get Dockerfiles with modified names
	DockerfileNames := getDockerfileName(client, repoFullName, username, path, recursionOption, ref)

	// Parse Dockerfiles
	if len(DockerfileNames) != 0 {

		for _, DockerfileName := range DockerfileNames {
			// Retrieve the repository file content
			fileContent, _, _, err := client.Repositories.GetContents(ctx, username, repoFullName, DockerfileName, nil)
			if err != nil {
				log.Println(err)
				continue // Skip this file and move to the next one
			}

			// Check if content is nil
			if fileContent == nil {
				log.Printf("Content is nil for file %s in repo %s\n", DockerfileName, repoFullName)
				continue // Skip this file and move to the next one
			}

			decodedContent, err := fileContent.GetContent()
			if err != nil {
				log.Println(err)
				continue // Skip this file and move to the next one
			}

			// Find Base image name
			container := findFROMLine(decodedContent)

			for containerName := range container {
				fmt.Println(container[containerName])
			}
		}
	}
}

// Function to get modified Dockerfile names
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

// Function to extract container name
func findFROMLine(content string) []string {
	lines := strings.Split(content, "\n")
	var FROMline []string

	// Regular expression to match "FROM" lines and capture the container name. Tweak this if necessary
	regex := regexp.MustCompile(`\b([a-zA-Z0-9\-.:/]+)(:latest)?\b`)
	regex2 := regexp.MustCompile(`\b(AS\b|platform\b|TARGETARCH\b)`)

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "FROM ") {
			line = strings.TrimPrefix(line, "FROM ")

			// Use regex2 to check if the line matches unwanted terms
			if !regex2.MatchString(line) {
				// Find matches in the line using the regular expression
				matches := regex.FindStringSubmatch(line)

				if len(matches) > 1 {
					// Extract the container name (group 1 in the regex match)
					containerName := matches[1]
					FROMline = append(FROMline, containerName)
				}
			}
		}
	}

	return FROMline
}

// Function to handle hitting rate limit
func handleRateLimit(err error, repoFullName string, iteration int) {
	log.Println(err)
	if e, ok := err.(*github.RateLimitError); ok {
		resetTime := e.Rate.Reset.Time
		sleepTime := time.Until(resetTime)
		log.Printf("Rate limit exceeded at %s. Try again in %s...\n", resetTime, sleepTime)
		log.Printf("When rate limit resets, restart from iteration %d. Repo to be scanned: %s\n", iteration, repoFullName)
		os.Exit(0)
	}

}
