# github-container-list

This script will crawl a Github Account or Organization and extract all of the base image names.

## Disclaimer

<b> Script will probably hit the Github rate limit</b>

This code is 50% written by chatGPT but is 100% gross.

## Usage

This script requires you to set ```GITHUB_TOKEN``` environment variable.

It accepts four arguments:

```
go run main.go <ORG_NAME> <BRANCH_NAME> <RECURSION_TRUE_OR_FALSE> <ITERATION_NUMBER>
```

If you have Dockerfiles in nested folders, set the recursion argument to '''true'''. Keep in mind that this will increase the API calls and hit the rate limit sooner.

If you want to set the iteration, you must also define the recursion setting.

<b> Suggested Usage: </B>

Basic:

```
go run main.go metalstormbass main  > out.txt
cat out.txt | sort | uniq
```

With Recursion: 
```
go run main.go metalstormbass main true  > out.txt
cat out.txt | sort | uniq
```

Continuiing after hitting rate limit:
```
go run main.go metalstormbass main true 160 >> out.txt
cat out.txt | sort | uniq
```