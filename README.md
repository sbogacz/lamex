# lamex

I couldn't come up with a good name, so I came up with a lame one instead: `LAMbdaEXecutor`.

## Motivation

We often write our system tests as Go tests. Sometimes these tests can also be pointed at either local or remote deployments.
The issue arises when our CI/CD pipeline needs to run system tests against services which are only accessible inside of our
VPC. We were `scp`ing test files to jump boxes and running them as a manual process. Automating it didn't feel much better.

So I was wondering whether it'd be possible to compile a Go test binary and run it in a Lambda. The answer was yes, but even
if the process returned `0`, Lambda would interpret that as a failed invocation, since it didn't receive the response it 
expected.

## How it works

This is a very simple approach to the problem. Take a newline separated text file for any commands you may want to run, parse
it, run the commands sequentially while piping their output to `stdout`, and return an error if any of them errored, else
return `nil`.

## How to use it

First compile the lamex binary to run on Lambda (or get it from the releases page)

```sh
cd $GOPATH/src/gitub.com/sbogacz/lamex
GOOS=linux GOARCH=amd64 go build -o lamex
```

Write your `commands.txt` file, e.g.

```sh
echo "./system-tests -test.v" > commands.txt
```

Zip it up with the binary you'd like to run, e.g. 

```sh
zip code.zip lamex yourCompiledBinary commands.txt
```

Upload it to Lambda, and make sure the Handler is set to `lamex` (or whatever name you gave it at compilation time)
