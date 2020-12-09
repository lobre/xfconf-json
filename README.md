# xfconf-json

An `xfconf-query` wrapper to apply configurations from a JSON file.

His big brother dconf has a way to import configurations from a file using:

```
dconf load / < config.ini
```

Unfortunately, xfconf does not support that out of the box. This only "programmable" way we have is by using the command line tool `xfconf-query`. But this is not really declarative, and I want to declaratively write my configuration. That's easier to keep them under version control for instance.

I know also that xfconf dumps its configurations in XML files under `./xfce4/xfconf/xfce-perchannel-xml`. But this is to me its internal, and I don't want to mess with xfconf's internals. I don't want to corrupt any file there, and configurations are not hot-reloaded when a file changes. I only want to use its public interface, which to me is `xfconf-query`.

That's why I created this simple `xfconf-query` wrapper that will call the corresponding `xfconf-query` commands, taking a JSON file as entry.

## Installation

This wrapper is developed in Go, and is not distributed to any package manager as for now. So you need Go installed and use:

```
go get -u github.com/lobre/xfconf-json
```

## Usage

```
Usage of ./xfconf-json:
  -bash
    	generate a bash script
  -bin string
    	xfconf-query binary (default "xfconf-query")
  -file string
    	json config file
```

So there are basically two modes depending on whether you use the `-bash` flag or not. If used, it will generate a bash script containing all the `xfconf-query` commands to apply. If not, it will directly apply them by automatically calling `xfconf-query` on your current system.

## Example

Here is a JSON file that you can take as example for the structure:

```
$ cat test.json

{
  "displays": {
    "/Default/eDP-1/Active": true,
    "/Schemes/Apply": "Default"
  },
  "xfce4-keyboard-shortcuts": {
    "/xfwm4/custom/<Primary><Shift>KP_Up": "tile_up_key"
  }
}
```

Here is what the bash script looks like:

```
$ ./xfconf-json -bash -file test.json

#!/usr/bin/env bash

# channel displays
xfconf-query --channel "displays" --property "/Default/eDP-1/Active" --create --type "bool" --set "true"
xfconf-query --channel "displays" --property "/Schemes/Apply" --create --type "string" --set "Default"

# channel xfce4-keyboard-shortcuts
xfconf-query --channel "xfce4-keyboard-shortcuts" --property "/xfwm4/custom/<Primary><Shift>KP_Up" --create --type "string" --set "tile_up_key"
```

And you can directly apply it without the `-bash` flag such as:

```
$ ./xfconf-json -file test.json
```
