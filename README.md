# Simple File Sync

Automatically back up your project files and sync file states across multiple devices.

## Features

- Synchronize file states across devices upon save.
- Comes with a simple CLI tool to manage files and directories.
- Intended for home LAN use but built with scaling capabilities.

## Tech

`sfs` uses a number of open source projects to work properly:

- [Chi] - Chi is a small, idiomatic and composable router for building HTTP services.
- [SQLite3] - A sqlite3 driver that conforms to the built-in database/sql interface.
- [envdecode] - envdecode is a Go package for populating structs from environment variables.
- [templ] - A language for writing HTML user interfaces in Go. 


## Installation

Simple File Sync requires [Go](https://go.dev/) v1.20+ to run.

Install the dependencies and initialize the project.

```sh
git clone https://github.com/JayDerderian/sfs.git

cd sfs
go mod install
go mod tidy
go build main.go -o sfs      # for linux/macOS
go build main.go -o sfs.exe  # for windows

# set executable to go path and test
cp sfs ~/go/bin              # change to where your go/bin file is located
export PATH:$PATH:$HOME/go/bin/sfs

sfs --version
```

Alternatively, you can use the `build.sh` script to build the project

## Configuration

Run `sfs setup` to run a first time setup of the project after compiling the source code. 

Use `sfs conf` to configure the the SFS client and server services after setup.

## License

MIT

**Free Software, Hell Yeah!**

[//]: # (These are reference links used in the body of this note and get stripped out when the markdown processor does its job. There is no need to format nicely because it shouldn't be seen. Thanks SO - http://stackoverflow.com/questions/4823468/store-comments-in-markdown-syntax)

   [Chi]: <https://pkg.go.dev/github.com/go-chi/chi>
   [SQLite3]: <https://pkg.go.dev/github.com/mattn/go-sqlite3>
   [envdecode]: <github.com/joeshaw/envdecode>
   [gotdotenv]: <github.com/joho/godotenv>
   [templ]: <https://github.com/a-h/templ>
