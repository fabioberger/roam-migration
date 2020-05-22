# Roam-migration

A command-line tool to convert [Roam Research](https://roamresearch.com/) exported files to [Org-roam](https://github.com/org-roam/org-roam) compatible markdown.

## Installation 

### Use a pre-built binary

Under the [Releases](https://github.com/fabioberger/roam-migration/releases) page we have pre-built binaries of the CLI available for most popular operating systems.

Download the relevant one for your system (e.g., darwin_amd64 for most MacOS users, etc...) and you're ready to go!

### Build from source

Building from source requires you to have [Golang](https://golang.org/) installed on your OS.

```
$ go get github.com/fabioberger/roam-migration
```

You should now have a CLI tool available called `roam-migration`

## Usage

First, go to Roam Research and click the three dots (`...`) in the top right corner and click "Export All". This will download a zip to your computer. Unzip this file by double clicking it. This will create a folder containing your Roam notes.

Then run the following:

```
$ roam-migration -p /path/to/roam-research-export
```

Replace the `/path/to` with where the roam-research export directory was saved on your machine.

If you downloaded a pre-built binary, you additionally need to replace `roam-migration`, with the path the the downloaded binary. 

For example, if you downloaded the pre-built binary on a Mac, it might look something like this:

```
$ ~/Downloads/roam-migration_darwin_amd64 -p ~/Downloads/Roam-Export-1590095488816
```

Happy hacking! If you run into any unexpected errors, please open an [issue](https://github.com/fabioberger/roam-migration/issues/new). 

## CLI arguments

`-h string` -- See help menu 

`-p string` -- Path to directory containing your Roam Research export
