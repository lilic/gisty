# gisty 

## What is gisty?

Command Line Interface tool for creating, editing and displaying [github gists](https://gist.github.com/).

## Installation

To install gisty just run:
```
go get github.com/lilic/gisty
```
Note: Works with go 1.6+

## Examples
For all available flags run:
```
gisty --help
```

To create a gist:
```
gisty --create --description="Description." --content="This is my gist." --filename="gist.md" --anon
```

Or create a gist by piping in a file as an input:
```
cat gist.md | gisty --create --filename="gist.md"
```

Get a gist by passing in a gist ID:
```
gisty --show="7ba6e7d22cbd168f6fbd010fda725105"
```

To edit a gist interactively just pass in the gist ID:
```
gisty --edit="7ba6e7d22cbd168f6fbd010fda725105"
```

List last 30 gists:
```
gisty --list
```
Note:
Make sure your ENV variable `$GITHUB_TOKEN` is set to the personal github access token.
