package main

import (
	"os"

	"Gtkmmcargo/builder"
)

/*
	A program called without parameters reads the configuration from the gtkmmcargo.cfg file.
	If the contents of the file seem to be correct, compilation and creation of the executable file are performed.
	If the file does not exist, the program exits.

	The parameter allowing to work with the configuration file is the '-cfg' flag.
	Example of use:
	gtkmmcargo -cfg template
		creates an empty cfg file on the disk in the current directory, in which the user can enter relevant data
	gtkmmcargo -cfg <cfg_file_name>
		allows the user to give the name of the configuration file (custom)
	gtkmmcargo -cfg (without parameter)
		displays the default flags for gtkmm
*/

func main() {
	cfgFile := ""

	args := parse(os.Args[1:])
	if v, ok := args["-cfg"]; ok {
		switch {
		case v == "":
			builder.PrintGtkmmFlags()
			return
		case v == "template":
			builder.NewEmpty().Save()
			return
		default:
			cfgFile = v
		}
	}

	if b := builder.New(cfgFile); b != nil {
		b.Build()
	}

	/*
		b := builder.New("/home/piotr/Projects/Gtkmm/Test/")
		b.AddFile("test.cc")
		//builder.PrintGtkmmFlags()
		ok := b.Build("testapp")
		if !ok {
			fmt.Println("failed")
		}

	*/
}

func parse(args []string) map[string]string {
	result := make(map[string]string)

	waitingKey := ""

	for _, item := range args {
		if item[0] == '-' {
			if waitingKey != "" {
				result[waitingKey] = ""
				waitingKey = ""
			}
			k, v := keyAndValue(item)
			switch {
			case k != "" && v != "":
				result[k] = v
			case k != "" && v == "":
				waitingKey = k
			}
		} else {
			if waitingKey != "" {
				result[waitingKey] = item
				waitingKey = ""
			}
		}
	}
	if waitingKey != "" {
		result[waitingKey] = ""
	}
	return result
}

func keyAndValue(text string) (string, string) {
	idx := -1

	i := 0
	for i < len(text) {
		if text[i] == '=' {
			idx = i
			break
		}
		i++
	}

	if idx > 0 && idx < (len(text)-1) {
		key := text[:idx]
		if key[0] == '-' {
			return text[:idx], text[idx+1:]
		}
		return "", ""
	}
	if text[0] == '-' {
		return text, ""
	}
	return "", text
}
