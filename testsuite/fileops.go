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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	return fmt.Errorf("Directory %s exists", dirName)
}

func fileShouldNotExist(fileName string) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil
	}

	return fmt.Errorf("File %s exists", fileName)
}

func getFileContent(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Cannot read file: %v", err)
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

	return CompareExpectedWithActualContains(expected, text)
}

func fileContentShouldNotContain(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return CompareExpectedWithActualNotContains(expected, text)
}

func fileContentShouldEqual(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return CompareExpectedWithActualEquals(expected, text)
}

func fileContentShouldNotEqual(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return CompareExpectedWithActualNotEquals(expected, text)
}

func fileContentShouldMatchRegex(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return CompareExpectedWithActualMatchesRegex(expected, text)
}

func fileContentShouldNotMatchRegex(filePath string, expected string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return CompareExpectedWithActualNotMatchesRegex(expected, text)
}

func fileContentIsInValidFormat(filePath string, format string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}

	return CheckFormat(format, text)
}
