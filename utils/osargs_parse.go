package utils

import (
	"fmt"
	"os"
)

func ParseOSArgs(namedArgsConfig map[string]bool) (map[string]string, []string, error) {
	args := os.Args
	positionalArgs := []string{}
	namedArgs := make(map[string]string)
	for index := 1; index < len(args); index++ {
		arg := args[index]
		isNamedArgument := arg[0] == '-'
		if !isNamedArgument {
			positionalArgs = append(positionalArgs, arg)
			continue
		}

		isFlagArgument, exists := namedArgsConfig[arg]
		if !exists {
			return nil, nil, fmt.Errorf(`unexpected named argument "%s"`, arg)
		}

		if isFlagArgument {
			namedArgs[arg] = "true"
			continue
		}

		index++
		if index == len(args) {
			return nil, nil, fmt.Errorf(`missing value for argument "%s"`, arg)
		}
		namedArgs[arg] = args[index]
	}
	return namedArgs, positionalArgs, nil
}
