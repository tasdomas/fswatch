# fswatch

`fswatch` is a simple command-line tool for watching filesystem changes and running a command
in response to them.

## Installing

Currently the only way to install `fswatch` is the `golang` toolkit:

``` shell
$ go install github.com/tasdomas/fswatch
```

## Usage

``` shell
fswatch [ --events create,write,remove,rename,chmod ] [ --command <command> ]
```

By default `fswatch` will listen for all events on a specified path (a directory or a file).
The events include:
 - `create`: creating a file or subdirectory
 - `write`: writing to a file
 - `remove`: removing of a file or subdirectory
 - `rename`: renaming a file or subdirectory
 - `chmod`: changing access permissions

Watching a directory will watch for changes to files and subdirectories directly inside it but will
not watch for changes inside subdirectories.

The change event's path and comma-separated list of event names can be passed to the command using
`{Path}` and `{Events}` placeholders:

``` shell
fswatch --command 'echo {Path} {Events}' ./
```
