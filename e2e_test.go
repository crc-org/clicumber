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

package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/code-ready/clicumber/testsuite"
	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
)

var opts = godog.Options{
	Format:              testsuite.GodogFormat,
	Paths:               strings.Split(testsuite.GodogPaths, ","),
	Tags:                testsuite.GodogTags,
	ShowStepDefinitions: testsuite.GodogShowStepDefinitions,
	StopOnFailure:       testsuite.GodogStopOnFailure,
	NoColors:            testsuite.GodogNoColors,
}

func init() {
	//testsuite.ParseFlags()

	godog.BindCommandLineFlags("godog.", &opts) // godog v0.11.0 and later
	//godog.BindCommandLineFlags("test.", &opts)  // godog v0.11.0 and later
	//godog.BindCommandLineFlags("-", &opts) // godog v0.11.0 and later
	testsuite.ParseFlags()
}

func TestMain(m *testing.M) {

	pflag.Parse()

	status := godog.TestSuite{
		Name:                 "clicumber",
		TestSuiteInitializer: testsuite.InitializeTestSuite,
		ScenarioInitializer:  testsuite.InitializeScenario,
		Options:              &opts,
	}.Run()

	os.Exit(status)
}
