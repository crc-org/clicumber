/*
Copyright (C) 2019 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testsuite

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func CompareExpectedWithActualContains(expected string, actual string) error {
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("Output did not match. Expected: '%s', Actual: '%s'", expected, actual)
	}

	return nil
}

func CompareExpectedWithActualNotContains(notexpected string, actual string) error {
	if strings.Contains(actual, notexpected) {
		return fmt.Errorf("Output did match. Not expected: '%s', Actual: '%s'", notexpected, actual)
	}

	return nil
}

func CompareExpectedWithActualEquals(expected string, actual string) error {
	if actual != expected {
		return fmt.Errorf("Output did not match. Expected: '%s', Actual: '%s'", expected, actual)
	}

	return nil
}

func CompareExpectedWithActualNotEquals(notexpected string, actual string) error {
	if actual == notexpected {
		return fmt.Errorf("Output did match. Not expected: '%s', Actual: '%s'", notexpected, actual)
	}

	return nil
}

func PerformRegexMatch(regex string, input string) (bool, error) {
	compRegex, err := regexp.Compile(regex)
	if err != nil {
		return false, fmt.Errorf("Expected value must be a valid regular expression statement: ", err)
	}

	return compRegex.MatchString(input), nil
}

func CompareExpectedWithActualMatchesRegex(expected string, actual string) error {
	matches, err := PerformRegexMatch(expected, actual)
	if err != nil {
		return err
	} else if !matches {
		return fmt.Errorf("Output did not match. Expected: '%s', Actual: '%s'", expected, actual)
	}

	return nil
}

func CompareExpectedWithActualNotMatchesRegex(notexpected string, actual string) error {
	matches, err := PerformRegexMatch(notexpected, actual)
	if err != nil {
		return err
	} else if matches {
		return fmt.Errorf("Output did match. Not expected: '%s', Actual: '%s'", notexpected, actual)
	}

	return nil
}

func CheckFormat(format string, actual string) error {
	actual = strings.TrimRight(actual, "\n")
	var err error
	switch format {
	case "URL":
		_, err = validateURL(actual)
	case "IP":
		_, err = validateIP(actual)
	case "IP with port number":
		_, err = validateIPWithPort(actual)
	case "YAML":
		_, err = validateYAML(actual)
	default:
		return fmt.Errorf("Format %s not implemented.", format)
	}

	return err
}

func validateIP(inputString string) (bool, error) {
	if net.ParseIP(inputString) == nil {
		return false, fmt.Errorf("IP address '%s' is not a valid IP address", inputString)
	}

	return true, nil
}

func validateURL(inputString string) (bool, error) {
	_, err := url.ParseRequestURI(inputString)
	if err != nil {
		return false, fmt.Errorf("URL '%s' is not an URL in valid format. Parsing error: %v", inputString, err)
	}

	return true, nil
}

func validateIPWithPort(inputString string) (bool, error) {
	split := strings.Split(inputString, ":")
	if len(split) != 2 {
		return false, fmt.Errorf("String '%s' does not contain one ':' separator", inputString)
	}
	if _, err := strconv.Atoi(split[1]); err != nil {
		return false, fmt.Errorf("Port must be an integer, in '%s' the port '%s' is not an integer. Conversion error: %v", inputString, split[1], err)
	}
	if net.ParseIP(split[0]) == nil {
		return false, fmt.Errorf("In '%s' the IP part '%s' is not a valid IP address", inputString, split[0])
	}

	return true, nil
}

func validateYAML(inputString string) (bool, error) {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(inputString), &m)
	if err != nil {
		return false, fmt.Errorf("Error unmarshaling YAML: %s. YAML='%s'", err, inputString)
	}

	return true, nil
}
