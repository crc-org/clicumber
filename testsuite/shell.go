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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/code-ready/clicumber/util"
)

const (
	exitCodeIdentifier = "exitCodeOfLastCommandInShell="

	bashExitCodeCheck       = "echo %v$?"
	fishExitCodeCheck       = "echo %v$status"
	tcshExitCodeCheck       = "echo %v$?"
	zshExitCodeCheck        = "echo %v$?"
	cmdExitCodeCheck        = "echo %v%%errorlevel%%"
	powershellExitCodeCheck = "echo %v$lastexitcode"
)

var (
	shell shellInstance
)

type shellInstance struct {
	startArgument    []string
	name             string
	checkExitCodeCmd string

	instance *exec.Cmd
	outbuf   bytes.Buffer
	errbuf   bytes.Buffer
	excbuf   bytes.Buffer

	outPipe io.ReadCloser
	errPipe io.ReadCloser
	inPipe  io.WriteCloser

	outScanner *bufio.Scanner
	errScanner *bufio.Scanner

	stdoutChannel   chan string
	stderrChannel   chan string
	exitCodeChannel chan string
}

func (shell shellInstance) getLastCmdOutput(stdType string) string {
	var returnValue string
	switch stdType {
	case "stdout":
		returnValue = shell.outbuf.String()
	case "stderr":
		returnValue = shell.errbuf.String()
	case "exitcode":
		returnValue = shell.excbuf.String()
	default:
		fmt.Printf("Field '%s' of shell's output is not supported. Only 'stdout', 'stderr' and 'exitcode' are supported.", stdType)
	}

	returnValue = strings.TrimSuffix(returnValue, "\n")

	return returnValue
}

func (shell *shellInstance) scanPipe(scanner *bufio.Scanner, buffer *bytes.Buffer, stdType string, channel chan string) {
	for scanner.Scan() {
		str := scanner.Text()
		util.LogMessage(stdType, str)

		if strings.Contains(str, exitCodeIdentifier) && !strings.Contains(str, shell.checkExitCodeCmd) {
			exitCode := strings.Split(str, "=")[1]
			shell.exitCodeChannel <- exitCode
		} else {
			buffer.WriteString(str + "\n")
		}
	}

	return
}

func (shell *shellInstance) configureTypeOfShell(shellName string) {
	switch shellName {
	case "bash":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(bashExitCodeCheck, exitCodeIdentifier)
	case "tcsh":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(tcshExitCodeCheck, exitCodeIdentifier)
	case "zsh":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(zshExitCodeCheck, exitCodeIdentifier)
	case "cmd":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(cmdExitCodeCheck, exitCodeIdentifier)
	case "powershell":
		shell.name = shellName
		shell.startArgument = []string{"-Command", "-"}
		shell.checkExitCodeCmd = fmt.Sprintf(powershellExitCodeCheck, exitCodeIdentifier)
	case "fish":
		fmt.Println("Fish shell is currently not supported by integration tests. Default shell for the OS will be used.")
		fallthrough
	default:
		if shell.name != "" {
			fmt.Printf("Shell %v is not supported, will set the default shell for the OS to be used.\n", shell.name)
		}
		switch runtime.GOOS {
		case "darwin", "linux":
			shell.name = "bash"
			shell.checkExitCodeCmd = fmt.Sprintf(bashExitCodeCheck, exitCodeIdentifier)
		case "windows":
			shell.name = "powershell"
			shell.startArgument = []string{"-Command", "-"}
			shell.checkExitCodeCmd = fmt.Sprintf(powershellExitCodeCheck, exitCodeIdentifier)
		}
	}

	return
}

func startHostShellInstance(shellName string) error {
	return shell.start(shellName)
}

func (shell *shellInstance) start(shellName string) error {
	var err error

	if shell.name == "" {
		shell.configureTypeOfShell(shellName)
	}
	shell.stdoutChannel = make(chan string)
	shell.stderrChannel = make(chan string)
	shell.exitCodeChannel = make(chan string)

	shell.instance = exec.Command(shell.name, shell.startArgument...)

	shell.outPipe, err = shell.instance.StdoutPipe()
	if err != nil {
		return err
	}

	shell.errPipe, err = shell.instance.StderrPipe()
	if err != nil {
		return err
	}

	shell.inPipe, err = shell.instance.StdinPipe()
	if err != nil {
		return err
	}

	shell.outScanner = bufio.NewScanner(shell.outPipe)
	shell.errScanner = bufio.NewScanner(shell.errPipe)

	go shell.scanPipe(shell.outScanner, &shell.outbuf, "stdout", shell.stdoutChannel)
	go shell.scanPipe(shell.errScanner, &shell.errbuf, "stderr", shell.stderrChannel)

	err = shell.instance.Start()
	if err != nil {
		return err
	}

	fmt.Printf("The %v instance has been started and will be used for testing.\n", shell.name)
	return err
}

func closeHostShellInstance() error {
	return shell.close()
}

func (shell *shellInstance) close() error {
	closingCmd := "exit\n"
	io.WriteString(shell.inPipe, closingCmd)
	err := shell.instance.Wait()
	if err != nil {
		fmt.Println("error closing shell instance:", err)
	}

	shell.instance = nil

	return err
}

func executeCommand(command string) error {
	if shell.instance == nil {
		return errors.New("shell instance is not started")
	}

	shell.outbuf.Reset()
	shell.errbuf.Reset()
	shell.excbuf.Reset()

	util.LogMessage(shell.name, command)

	_, err := io.WriteString(shell.inPipe, command+"\n")
	if err != nil {
		return err
	}

	_, err = shell.inPipe.Write([]byte(shell.checkExitCodeCmd + "\n"))
	if err != nil {
		return err
	}

	exitCode := <-shell.exitCodeChannel
	shell.excbuf.WriteString(exitCode)

	return err
}

func executeCommandSucceedsOrFails(command string, expectedResult string) error {
	err := executeCommand(command)
	if err != nil {
		return err
	}

	exitCode := shell.excbuf.String()

	if expectedResult == "succeeds" && exitCode != "0" {
		err = fmt.Errorf("command '%s', expected to succeed, exited with exit code: %s", command, exitCode)
	}
	if expectedResult == "fails" && exitCode == "0" {
		err = fmt.Errorf("command '%s', expected to fail, exited with exit code: %s", command, exitCode)
	}

	return err
}

func executeCommandWithRetry(retryCount int, retryTime string, command string, expected string) error {
	var exitCode, stdout string
	retryDuration, err := time.ParseDuration(retryTime)
	if err != nil {
		return err
	}

	for i := 0; i < retryCount; i++ {
		err := executeCommand(command)
		exitCode, stdout := shell.excbuf.String(), shell.outbuf.String()
		if err == nil && exitCode == "0" && strings.Contains(stdout, expected) {
			return nil
		}
		time.Sleep(retryDuration)
	}

	return fmt.Errorf("command '%s', Expected: exitCode 0, stdout %s, Actual: exitCode %s, stdout %s", command, expected, exitCode, stdout)
}

func executeStdoutLineByLine() error {
	var err error
	stdout := shell.getLastCmdOutput("stdout")
	commandArray := strings.Split(stdout, "\n")
	for index := range commandArray {
		if !strings.Contains(commandArray[index], exitCodeIdentifier) {
			err = executeCommand(commandArray[index])
		}
	}

	return err
}

func commandReturnShouldContain(commandField string, expected string) error {
	return compareExpectedWithActualContains(expected, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldContainContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualContains(expected.Content, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldNotContain(commandField string, notexpected string) error {
	return compareExpectedWithActualNotContains(notexpected, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldNotContainContent(commandField string, notexpected *gherkin.DocString) error {
	return compareExpectedWithActualNotContains(notexpected.Content, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldBeEmpty(commandField string) error {
	return compareExpectedWithActualEquals("", shell.getLastCmdOutput(commandField))
}

func commandReturnShouldNotBeEmpty(commandField string) error {
	return compareExpectedWithActualNotEquals("", shell.getLastCmdOutput(commandField))
}

func commandReturnShouldEqual(commandField string, expected string) error {
	return compareExpectedWithActualEquals(expected, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldEqualContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualEquals(expected.Content, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldNotEqual(commandField string, expected string) error {
	return compareExpectedWithActualNotEquals(expected, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldNotEqualContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualNotEquals(expected.Content, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldMatch(commandField string, expected string) error {
	return compareExpectedWithActualMatchesRegex(expected, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldMatchContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualMatchesRegex(expected.Content, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldNotMatch(commandField string, expected string) error {
	return compareExpectedWithActualNotMatchesRegex(expected, shell.getLastCmdOutput(commandField))
}

func commandReturnShouldNotMatchContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualNotMatchesRegex(expected.Content, shell.getLastCmdOutput(commandField))
}

func shouldBeInValidFormat(commandField string, format string) error {
	return checkFormat(format, shell.getLastCmdOutput(commandField))
}

func setScenarioVariableExecutingCommand(variableName string, command string) error {
	err := executeCommand(command)
	if err != nil {
		return err
	}

	commandFailed := (shell.getLastCmdOutput("exitcode") != "0" || len(shell.getLastCmdOutput("stderr")) != 0)
	if commandFailed {
		return fmt.Errorf("command '%v' did not execute successfully. cmdExit: %v, cmdErr: %v",
			command,
			shell.getLastCmdOutput("exitcode"),
			shell.getLastCmdOutput("stderr"))
	}

	stdout := shell.getLastCmdOutput("stdout")
	util.SetScenarioVariable(variableName, strings.TrimSpace(stdout))

	return nil
}
