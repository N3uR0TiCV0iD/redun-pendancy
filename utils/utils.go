package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Not a fan of Go's ternary syntax :P
func TernarySelect[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

func OpenAndProcessFileLines(filePath string, handler func(string) error) error {
	return OpenAndProcessFile(filePath, func(file *os.File) error {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			err := handler(line)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func OpenAndProcessFile(filePath string, handler func(*os.File) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	err = handler(file)
	return err
}

func ValueOrDefault(value string, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

func TrimPrefixIfMatch(input string, prefix string) (string, bool) {
	if strings.HasPrefix(input, prefix) {
		return input[len(prefix):], true
	}
	return input, false
}

// Removes any special characters from a version string
func NormalizeVersion(version string) string {
	regex := regexp.MustCompile(`[\[\]]`)
	return regex.ReplaceAllString(version, "")
}

func IsVersionHigher(left string, right string) bool {
	result, err := CompareVersions(left, right)
	if err != nil {
		log.Printf("[Warning] Failed to compare versions '%s' and '%s': %v", left, right, err)
		return false
	}
	return result > 0
}

func CompareVersions(left string, right string) (int, error) {
	if left == right {
		return 0, nil
	}

	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	maxCompares := min(len(leftParts), len(rightParts))

	for index := 0; index < maxCompares; index++ {
		leftNum, err := strconv.Atoi(leftParts[index])
		if err != nil {
			return 0xDEADC0DE, fmt.Errorf("%w at segment %d in left version '%s'", err, index, left)
		}

		rightNum, err := strconv.Atoi(rightParts[index])
		if err != nil {
			return 0xDEADC0DE, fmt.Errorf("%w at segment %d in right version '%s'", err, index, right)
		}

		if leftNum > rightNum {
			return 1, nil
		} else if leftNum < rightNum {
			return -1, nil
		}
	}

	if len(leftParts) > len(rightParts) {
		return 1, nil
	} else if len(leftParts) < len(rightParts) {
		return -1, nil
	}

	return 0, nil
}

func BuildPackageKey(name string, version string, framework string) string {
	return fmt.Sprintf("%s=%s;%s", name, version, framework)
}

func TimeFunction(functionName string, function func()) {
	start := time.Now()
	function()
	elapsed := time.Since(start)
	log.Printf(`Function "%s()" executed in: %s`, functionName, elapsed)
}
