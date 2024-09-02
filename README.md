# GoGrep

This repo contains the program I wrote for the *["Build Your Own grep" Challenge](https://app.codecrafters.io/courses/grep/overview)*. on [codecrafters.io](https://codecrafters.io). I've been really enjoying both CodeCrafters and writing Go, so as yet another month has come to pass and they're again offering a free challenge, how can I say no? (here's a [link](https://github.com/jasonflorentino/go-http-server) to last month's)

And yet.

"Make your own `grep`," they said; "Regex is *fun*," they said. All *I* have to say is that I'm thankful CodeCrafters kept it nice and simple for what tests they'd run. I hate to let down Rob and Brian, but summer is almost over, and I'd like to get off Regex's wild ride.

— Jason, August 2024

### From the CodeCrafters overview:

[Regular expressions](https://en.wikipedia.org/wiki/Regular_expression)
(Regexes, for short) are patterns used to match character combinations in
strings. [`grep`](https://en.wikipedia.org/wiki/Grep) is a CLI tool for
searching using Regexes.

In this challenge you'll build your own implementation of `grep`. Along the way
we'll learn about Regex syntax, how parsers/lexers work, and how regular
expressions are evaluated.

# Usage

- Run the program using `go run`:
  ```
  go run src/main.go
  ```
- Input is taken from `stdin`:
  ```
  echo caaaats | go run src/main.go
  ```
- Specify a pattern to match against with the `-E` flag:
  ```
  echo caaaats | go run src/main.go -E ca+t
  ```
- You can view very ugly debug information with the `--debug` flag:
  ```
  echo caaaats | go run src/main.go -E ca+t --debug
  ```

## Features
- Matches literal characters
- Matches digits with `\d`
- Matches alphanumeric characters with `\w`
- Maches one of a set of characters with `[abc]`
- Negatively maches one of a set of characters with `[^abc]`
- Matches the start of a string with `^`
- Matches the end of a string with `$`
- Matches one or more times with `+`
- Matches zero or more times with `?`
- Supports alternation with `|` like in `(cat|dog)`
- Supports backreferences with `\1` like in `super(man|woman) and bat\1`
- Is otherwise very limited and buggy (probably)

## Automated Tests
While working through the challenge stages, I maintained a test script to help verify new additions and catch regressions before submitting:
```bash
./test.sh
```
