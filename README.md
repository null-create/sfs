# Simple File Sync

Automatically back up your project files and sync file states across multiple devices.

## Features

- Synchronize file states across devices upon save.
- Comes with a simple CLI tool to manage files and directories.
- Comes with an optional web interface
- Intended for home LAN use but built to scale.

## Demos

## Tech

`sfs` was built using a number of open source projects:

- [fsnotify] - Cross-platform filesystem notifications for Go. 
- [Chi] - Chi is a small, idiomatic and composable router for building HTTP services.
- [SQLite3] - A sqlite3 driver that conforms to the built-in database/sql interface.
- [cobra] - Cobra is a library for creating powerful modern CLI applications.
- [viper] - Go configuration with fangs 
- [envdecode] - envdecode is a Go package for populating structs from environment variables.


## Installation

Simple File Sync requires [Go](https://go.dev/) v1.21+ to run.

Download and the latest binary from GitHub

```bash
curl -LO https://github.com/JayDerderian/sfs/releases/latest
```

Build using source code and supplied `build.sh` script

```sh
git clone https://github.com/JayDerderian/sfs.git
cd sfs
chmod +x build.sh
./build.sh

# set executable to go path and test
export PATH:$PATH:$PWD/sfs/bin

sfs --version
```

Alternatively, you can use the `build.sh` script to build the project.

`sfs` can be ran in a Docker container. Run the following command to  build the project image in the project directory:

```bash
docker build -t sfs .
```

## Configuration

See docs/CONFIGURATION.md for more in-depth information about how to configure the project for home LAN use, as well as other kinds of set up types.

- Run `sfs setup` to run the **first time setup** of the project after compiling the source code. 
- Use `sfs conf` to configure the the SFS client and server services **after** setup.

If you want to manually configure the SFS client and server services, you will 
need to modify the yaml file under pkg/configs. This file is ready at startup and be used to set up the necessary environment variables at runtime.

## License

MIT

**Free Software, Hell Yeah!**

[//]: # (These are reference links used in the body of this note and get stripped out when the markdown processor does its job. There is no need to format nicely because it shouldn't be seen. Thanks SO - http://stackoverflow.com/questions/4823468/store-comments-in-markdown-syntax)

   [Chi]: <https://pkg.go.dev/github.com/go-chi/chi>
   [fsnotify]: <https://github.com/fsnotify/fsnotify>
   [SQLite3]: <https://pkg.go.dev/github.com/mattn/go-sqlite3>
   [cobra]: <https://github.com/spf13/cobra>
   [viper]: <https://github.com/spf13/viper>
   [envdecode]: <github.com/joeshaw/envdecode>
   [gotdotenv]: <github.com/joho/godotenv>
   [templ]: <https://github.com/a-h/templ>
