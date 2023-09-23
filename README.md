# github-container-list

This script will crawl a Github Account or Organization and extract all of the container image names.

## Disclaimer

<b> This will probably hit the Github rate limit - It will handle it gracefully and wait to continue </b>

This code is 50% written by chatGPT but is 100% gross.

## Usage

This script requires you to set ```GITHUB_TOKEN``` environment variable.

It accepts three arguments:

```
go run main.go <ORG_NAME> <BRANCH_NAME> <RECURSION_TRUE_OR_FALSE>
```

If you have Dockerfiles in nested folders, set the last argument to '''true'''. Keep in mind that this will increase the API calls and hit the rate limit sooner.

<b> Suggested Usage: </B>
```
go run main.go metalstormbass main false  > out.txt
cat out.txt | sort | uniq
```

