# github-container-list

This script will crawl a Github Account or Organization and extract all of the container image names.

## Disclaimer

<b> This will probably hit the Github rate limit </B>

This code is half written by chat GPT. It is 100% gross.

## Usage

This script requires you to set ```GITHUB_TOKEN``` environment variable.

It accepts three arguments:

```
go run main.go <ORG_NAME> <BRANCH_NAME> <RECURSION_TRUE_OR_FALSE>
```

<b> Suggested Usage: </B>
```
go run main.go metalstormbass main false  > out.txt
cat out.txt | sort | uniq
```

