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
	"os"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"

	"github.com/agajdosi/clicumber/util"
)

var (
	testDir         string
	testRunDir      string
	testResultsDir  string
	testDefaultHome string
	testWithShell   string

	GodogFormat              string
	GodogTags                string
	GodogShowStepDefinitions bool
	GodogStopOnFailure       bool
	GodogNoColors            bool
	GodogPaths               string
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {
	// Executing commands
	s.Step(`^executing "(.*)"$`,
		executeCommand)
	s.Step(`^executing "(.*)" (succeeds|fails)$`,
		executeCommandSucceedsOrFails)

	// Command output verification
	s.Step(`^(stdout|stderr|exitcode) (?:should contain|contains) "(.*)"$`,
		commandReturnShouldContain)
	s.Step(`^(stdout|stderr|exitcode) (?:should contain|contains)$`,
		commandReturnShouldContainContent)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not contain "(.*)"$`,
		commandReturnShouldNotContain)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does not) contain$`,
		commandReturnShouldNotContainContent)

	s.Step(`^(stdout|stderr|exitcode) (?:should equal|equals) "(.*)"$`,
		commandReturnShouldEqual)
	s.Step(`^(stdout|stderr|exitcode) (?:should equal|equals)$`,
		commandReturnShouldEqualContent)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not equal "(.*)"$`,
		commandReturnShouldNotEqual)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not equal$`,
		commandReturnShouldNotEqualContent)

	s.Step(`^(stdout|stderr|exitcode) (?:should match|matches) "(.*)"$`,
		commandReturnShouldMatch)
	s.Step(`^(stdout|stderr|exitcode) (?:should match|matches)`,
		commandReturnShouldMatchContent)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not match "(.*)"$`,
		commandReturnShouldNotMatch)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not match`,
		commandReturnShouldNotMatchContent)

	s.Step(`^(stdout|stderr|exitcode) (?:should be|is) empty$`,
		commandReturnShouldBeEmpty)
	s.Step(`^(stdout|stderr|exitcode) (?:should not be|is not) empty$`,
		commandReturnShouldNotBeEmpty)

	s.Step(`^(stdout|stderr|exitcode) (?:should be|is) valid "([^"]*)"$`,
		shouldBeInValidFormat)

	// Command output and execution: extra steps
	s.Step(`^with up to "(\d*)" retries with wait period of "(\d*(?:ms|s|m))" command "(.*)" output (?:should contain|contains) "(.*)"$`,
		executeCommandWithRetry)
	s.Step(`^evaluating stdout of the previous command succeeds$`,
		executeStdoutLineByLine)

	// Scenario variables
	// allows to set a scenario variable to the output values of minishift and oc commands
	// and then refer to it by $(NAME_OF_VARIABLE) directly in the text of feature file
	s.Step(`^setting scenario variable "(.*)" to the stdout from executing "(.*)"$`,
		setScenarioVariableExecutingCommand)

	// Filesystem operations
	s.Step(`^creating directory "([^"]*)" succeeds$`,
		createDirectory)
	s.Step(`^creating file "([^"]*)" succeeds$`,
		createFile)
	s.Step(`^deleting directory "([^"]*)" succeeds$`,
		deleteDirectory)
	s.Step(`^deleting file "([^"]*)" succeeds$`,
		deleteFile)
	s.Step(`^directory "([^"]*)" should not exist$`,
		directoryShouldNotExist)
	s.Step(`^file "([^"]*)" should not exist$`,
		fileShouldNotExist)
	s.Step(`^file from "(.*)" is downloaded into location "(.*)"$`,
		downloadFileIntoLocation)
	s.Step(`^writing text "([^"]*)" to file "([^"]*)" succeeds$`,
		writeToFile)

	// File content checks
	s.Step(`^content of file "([^"]*)" should contain "([^"]*)"$`,
		fileContentShouldContain)
	s.Step(`^content of file "([^"]*)" should not contain "([^"]*)"$`,
		fileContentShouldNotContain)
	s.Step(`^content of file "([^"]*)" should equal "([^"]*)"$`,
		fileContentShouldEqual)
	s.Step(`^content of file "([^"]*)" should not equal "([^"]*)"$`,
		fileContentShouldNotEqual)
	s.Step(`^content of file "([^"]*)" should match "([^"]*)"$`,
		fileContentShouldMatchRegex)
	s.Step(`^content of file "([^"]*)" should not match "([^"]*)"$`,
		fileContentShouldNotMatchRegex)
	s.Step(`^content of file "([^"]*)" (?:should be|is) valid "([^"]*)"$`,
		fileContentIsInValidFormat)

	s.BeforeSuite(func() {
		err := prepareForIntegrationTest()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})

	s.BeforeFeature(func(this *gherkin.Feature) {
		util.LogMessage("info", fmt.Sprintf("----- Feature: %s -----", this.Name))
		startHostShellInstance(testWithShell)
		util.ClearScenarioVariables()
		err := cleanTestRunDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})

	s.BeforeScenario(func(this interface{}) {
		switch this.(type) {
		case *gherkin.Scenario:
			scenario := *this.(*gherkin.Scenario)
			util.LogMessage("info", fmt.Sprintf("----- Scenario: %s -----", scenario.ScenarioDefinition.Name))
		case *gherkin.ScenarioOutline:
			scenario := *this.(*gherkin.ScenarioOutline)
			util.LogMessage("info", fmt.Sprintf("----- Scenario Outline: %s -----", scenario.ScenarioDefinition.Name))
		}
	})

	s.BeforeStep(func(this *gherkin.Step) {
		this.Text = util.ProcessScenarioVariables(this.Text)
		switch v := this.Argument.(type) {
		case *gherkin.DocString:
			v.Content = util.ProcessScenarioVariables(v.Content)
		}
	})

	s.AfterScenario(func(interface{}, error) {
	})

	s.AfterFeature(func(this *gherkin.Feature) {
		util.LogMessage("info", "----- Cleaning after feature -----")
		closeHostShellInstance()
	})

	s.AfterSuite(func() {
		util.LogMessage("info", "----- Cleaning Up -----")
		err := util.CloseLog()
		if err != nil {
			fmt.Println("Error closing the log:", err)
		}
	})
}
