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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func createDirectory(dirName string) error {
	return os.MkdirAll(dirName, 0777)
}

func deleteDirectory(dirName string) error {
	return os.RemoveAll(dirName)
}

func deleteFile(fileName string) error {
	return os.RemoveAll(fileName)
}

func directoryShouldNotExist(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		return nil
	}

	return fmt.Errorf("directory %s exists", dirName)
}

func fileShouldNotExist(fileName string) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil
	}

	return fmt.Errorf("file %s exists", fileName)
}

func fileExist(fileName string) error {

	_, err := os.Stat(fileName)

	if os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exists, error: %v ", fileName, err)
	} else if err != nil {
		return fmt.Errorf("file %s neither exists nor doesn't exist, error: %v", fileName, err)
	}
	return nil // == err
}

func getFileContent(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %v", err)
	}

	return string(data), nil
}

func createFile(fileName string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func writeToFile(text string, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(text)
	if err != nil {
		return err
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	return nil
}

func downloadFileIntoLocation(downloadURL string, destinationFolder string) error {
	destinationFolder = filepath.Join(testRunDir, destinationFolder)
	err := os.MkdirAll(destinationFolder, os.ModePerm)
	if err != nil {
		return err
	}

	slice := strings.Split(downloadURL, "/")
	fileName := slice[len(slice)-1]
	filePath := filepath.Join(destinationFolder, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func fileContentShouldContain(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return compareExpectedWithActualContains(expected, text)
}

func fileContentShouldNotContain(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return compareExpectedWithActualNotContains(expected, text)
}

func fileContentShouldEqual(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return compareExpectedWithActualEquals(expected, text)
}

func fileContentShouldNotEqual(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return compareExpectedWithActualNotEquals(expected, text)
}

func fileContentShouldMatchRegex(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return compareExpectedWithActualMatchesRegex(expected, text)
}

func fileContentShouldNotMatchRegex(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return compareExpectedWithActualNotMatchesRegex(expected, text)
}

func fileContentIsInValidFormat(filePath string, format string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return checkFormat(format, text)
}

// --------------- CONFIG functions from minishift

func configFileContainsKeyMatchingValue(format string, configPath string, condition string, keyPath string, expectedValue string) error {
	config, err := getFileContent(configPath)
	if err != nil {
		return err
	}

	keyValue, err := getConfigKeyValue([]byte(config), format, keyPath)
	if err != nil {
		return err
	}

	matches, err := performRegexMatch(expectedValue, keyValue)
	if err != nil {
		return err
	} else if (condition == "contains") && !matches {
		return fmt.Errorf("For key '%s' config contains unexpected value '%s'", keyPath, keyValue)
	} else if (condition == "does not contain") && matches {
		return fmt.Errorf("For key '%s' config contains value '%s', which it should not contain", keyPath, keyValue)
	}

	return nil
}

func configFileContainsKey(format string, configPath string, condition string, keyPath string) error {
	config, err := getFileContent(configPath)
	if err != nil {
		return err
	}

	keyValue, err := getConfigKeyValue([]byte(config), format, keyPath)
	if err != nil {
		return err
	}

	if (condition == "contains") && (keyValue == "<nil>") {
		return fmt.Errorf("Config does not contain any value for key %s", keyPath)
	} else if (condition == "does not contain") && (keyValue != "<nil>") {
		return fmt.Errorf("Config contains key %s with assigned value: %s", keyPath, keyValue)
	}

	return nil
}

func getConfigKeyValue(configData []byte, format string, keyPath string) (string, error) {
	var err error
	var keyValue string
	var values map[string]interface{}

	if format == "JSON" {
		err = json.Unmarshal(configData, &values)
		if err != nil {
			return "", fmt.Errorf("Error unmarshaling JSON: %s", err)
		}
	} else if format == "YAML" {
		err = yaml.Unmarshal(configData, &values)
		if err != nil {
			return "", fmt.Errorf("Error unmarshaling YAML: %s", err)
		}
	}

	keyPathArray := strings.Split(keyPath, ".")
	for _, element := range keyPathArray {
		switch value := values[element].(type) {
		case map[string]interface{}:
			values = value
		case map[interface{}]interface{}:
			retypedValue := make(map[string]interface{})
			for x := range value {
				retypedValue[x.(string)] = value[x]
			}
			values = retypedValue
		case []interface{}, nil, string, int, float64, bool:
			keyValue = fmt.Sprintf("%v", value)
		default:
			return "", errors.New("Unexpected type in config file, type not supported.")
		}
	}
	return keyValue, nil
}
