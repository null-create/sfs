# Simple File Sync

### NOTE: Under active development! Not all components are fully functional yet.

[![Build Status](https://travis-ci.org/joemccann/dillinger.svg?branch=master)](https://github.com/JayDerderian/sfs/tree/mai)

Automatically back up your project files and sync file states across multiple devices.

## Features

- Syncrhonize file states across devices upon save.
- Manage a local and remote file system via a simple browser interface.
- Comes with a robust CLI tool to manage files and directories.
- Intended for home use but built with scaling capabilities.


## Tech

`SFS` uses a number of open source projects to work properly:

- [Chi] - Chi is a small, idiomatic and composable router for building HTTP services.
- [SQLite3] - A sqlite3 driver that conforms to the built-in database/sql interface



## Installation

Simple File Sync requires [Go](https://go.dev/) v1.19+ to run.

Install the dependencies and initialize the project.

```sh
cd path/to/sfs
go mod install
go mod tidy

go build main.go -o sfs      # for linux/macOS
go build main.go -o sfs.exe  # for windows

# set executable to go path and test
cp sfs ~/go/bin 
sfs --version
```

## Docker

Dillinger is very easy to install and deploy in a Docker container.

By default, the Docker will expose port 8080, so change this within the
Dockerfile if necessary. When ready, simply use the Dockerfile to
build the image.

```sh
cd sfs
docker build -t <youruser>/sfs:${package.json.version} .
```

This will create the `sfs` image and pull in the necessary dependencies.
Be sure to swap out `${package.json.version}` with the actual
version of `Simple File Sync`.

Once done, run the Docker image and map the port to whatever you wish on
your host. In this example, we simply map port 8000 of the host to
port 8080 of the Docker (or whatever port was exposed in the Dockerfile):

```sh
docker run -d -p 8000:8080 --restart=always  --name=sfs <youruser>/sfs:${package.json.version}
```


Verify the deployment by navigating to your server address in
your preferred browser.

```sh
127.0.0.1:8000
```

## License

MIT

**Free Software, Hell Yeah!**

[//]: # (These are reference links used in the body of this note and get stripped out when the markdown processor does its job. There is no need to format nicely because it shouldn't be seen. Thanks SO - http://stackoverflow.com/questions/4823468/store-comments-in-markdown-syntax)


   [Chi]: <https://pkg.go.dev/github.com/go-chi/chi>
   [SQLite3]: <https://pkg.go.dev/github.com/mattn/go-sqlite3>

   [PlDb]: <https://github.com/joemccann/dillinger/tree/master/plugins/dropbox/README.md>
   [PlGh]: <https://github.com/joemccann/dillinger/tree/master/plugins/github/README.md>
   [PlGd]: <https://github.com/joemccann/dillinger/tree/master/plugins/googledrive/README.md>
   [PlOd]: <https://github.com/joemccann/dillinger/tree/master/plugins/onedrive/README.md>
   [PlMe]: <https://github.com/joemccann/dillinger/tree/master/plugins/medium/README.md>
   [PlGa]: <https://github.com/RahulHP/dillinger/blob/master/plugins/googleanalytics/README.md>
