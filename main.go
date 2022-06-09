package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	in := flag.String("file", "", "json config file")
	bin := flag.String("bin", "xfconf-query", "xfconf-query binary")
	bash := flag.Bool("bash", false, "generate a bash script")
	flag.Parse()

	if *in == "" {
		return errors.New("file parameter not provided")
	}

	f, err := os.Open(*in)
	if err != nil {
		return err
	}

	conf, err := parseConfig(f)
	if err != nil {
		return errors.New("cannot parse configuration")
	}

	if *bash {
		fmt.Println(conf.toBash(*bin))
		return nil
	}

	return conf.apply(*bin)
}

type channel map[string]interface{}
type config map[string]channel

// parseConfig creates a config from a json input.
func parseConfig(r io.Reader) (*config, error) {
	var conf config

	dec := json.NewDecoder(r)

	err := dec.Decode(&conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

// toBash generates a bash script containing all the calls
// to xfconf-query corresponding to what was described in
// the original json file. It can be used if one does not
// want to directly call apply(), or if the generation should
// happen on a different machine without xfconf-query installed.
func (c *config) toBash(bin string) string {
	var b strings.Builder

	if len(*c) > 0 {
		b.WriteString("#!/usr/bin/env bash")
	}

	for channel, props := range *c {
		fmt.Fprintf(&b, "\n\n# channel %s", channel)

		for prop, v := range props {
			args := args(channel, prop, v, true)
			fmt.Fprintf(&b, "\n%s %s", bin, strings.Join(args, " "))
		}
	}

	return b.String()
}

// apply wraps command calls to xfconf-query using os/exec and
// will apply the configurations to the current system.
func (c *config) apply(bin string) error {
	xfconfBin, err := exec.LookPath(bin)
	if err != nil {
		return fmt.Errorf("%s is not found on your system", bin)
	}

	for channel, props := range *c {
		for prop, v := range props {
			args := args(channel, prop, v, false)
			cmd := exec.Command(xfconfBin, args...)

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				return errors.New(stderr.String())
			}
		}
	}

	return nil
}

// args generates the correct list of arguments that will be used
// by the xfconf-query tool.
func args(channel string, prop string, v interface{}, escape bool) []string {
	format := func(s string) string {
		if escape {
			return fmt.Sprintf("%q", s)
		}
		return s
	}

	// make sure property starts with slash
	prop = fmt.Sprintf("/%s", strings.Trim(prop, "/"))

	var args []string
	args = append(args, "--channel", format(channel))
	args = append(args, "--property", format(prop))
	args = append(args, "--create")

	// if array, add --type and --set flags multiple times
	arr, isArr := v.([]interface{})
	if isArr {
		// force array even if one element
		args = append(args, "--force-array")

		for _, vv := range arr {
			args = append(args, "--type", format(xfconfType(vv)))
			args = append(args, "--set", format(fmt.Sprintf("%v", vv)))
		}
	} else {
		args = append(args, "--type", format(xfconfType(v)))
		args = append(args, "--set", format(fmt.Sprintf("%v", v)))
	}

	return args
}

// xfconfType returns a string containing the
// xfconf type corresponding to a go type.
//
// The list of xfconf types is: int, uint, bool,
// float, double, string.
//
// Only int, bool, double and string are used here.
//
// xfconf also has arrays, but they are not handled
// by this function.
func xfconfType(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case float64:
		// detect as int if it is a round number
		if v.(float64) == math.Trunc(v.(float64)) {
			return "int"
		}
		return "double"
	case string:
		return "string"
	case bool:
		return "bool"
	default:
		return ""
	}
}
