# Simple File Sync

### NOTE: Under active development! Not all components are fully functional yet.

[![Build Status](https://travis-ci.org/joemccann/dillinger.svg?branch=master)](https://github.com/JayDerderian/sfs/tree/main)

Automatically back up your project files and sync file states across multiple devices.

## Features

- Synchronize file states across devices upon save.
- Manage a local and remote file system via a simple browser interface.
- Comes with a robust CLI tool to manage files and directories.
- Intended for home LAN use but built with scaling capabilities.


## Tech

`SFS` uses a number of open source projects to work properly:

- [Chi] - Chi is a small, idiomatic and composable router for building HTTP services.
- [SQLite3] - A sqlite3 driver that conforms to the built-in database/sql interface.
- [godotenv] - A Go port of Ruby's dotenv library (Loads environment variables from .env files)
- [envdecode] - envdecode is a Go package for populating structs from environment variables.


## Installation

Simple File Sync requires [Go](https://go.dev/) v1.21+ to run.

Install the dependencies and initialize the project.

```sh
cd path/to/sfs
go mod install
go mod tidy

go build main.go -o sfs      # for linux/macOS
go build main.go -o sfs.exe  # for windows

# set executable to go path and test
cp sfs ~/go/bin              # change to where your go/bin file is located
sfs --version
```

## License

MIT

**Free Software, Hell Yeah!**

[//]: # (These are reference links used in the body of this note and get stripped out when the markdown processor does its job. There is no need to format nicely because it shouldn't be seen. Thanks SO - http://stackoverflow.com/questions/4823468/store-comments-in-markdown-syntax)

   [Chi]: <https://pkg.go.dev/github.com/go-chi/chi>
   [SQLite3]: <https://pkg.go.dev/github.com/mattn/go-sqlite3>
   [envdecode]: <github.com/joeshaw/envdecode>
   [gotdotenv]: <github.com/joho/godotenv>
