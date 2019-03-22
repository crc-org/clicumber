Feature: Testsuite test
Quentin check whether their testsuite works properly.

  Scenario: Contains
     When executing "go help" succeeds
     Then stdout should contain
     """
     Go is a tool for managing Go source code.
     """

  Scenario: Not Contains
     When executing "go help" succeeds
     Then stdout should not contain "Error"

  Scenario: Equals
     When executing "go help" succeeds
     Then exitcode should equal "0"

  Scenario: Not Equals
     When executing "go notexist" fails
     Then exitcode should not equal "0"

  Scenario: Matches
     When executing "go version" succeeds
     Then stdout should match "go1\.\d+\.\d+"

  Scenario: Not Matches
     When executing "go version" succeeds
     Then stdout should not match "Local \d\.\d OpenShift clusters"

  Scenario: Is Empty
     When executing "go version" succeeds
     Then stderr should be empty

  Scenario: Is Not Empty
     When executing "go version" succeeds
     Then stdout should not be empty

  Scenario: Scenario Variables
     When setting scenario variable "VAR" to the stdout from executing "go version"
      And executing "echo $(VAR)" succeeds
     Then stdout should contain "go version"

   Scenario: Create Directory and Files
      When creating directory "newdir" succeeds
       And creating file "newdir/newfile" succeeds
       And file from "https://google.com" is downloaded into location "newdir"
      
   Scenario: File Content Checks
      When writing text "192.168.15.17" to file "newdir/newfile" succeeds
      Then content of file "newdir/newfile" should contain "168"
       And content of file "newdir/newfile" should not contain "512"
       And content of file "newdir/newfile" should equal "192.168.15.17"
       And content of file "newdir/newfile" should not equal "192.168.15.512"
       And content of file "newdir/newfile" should match "192\.168\.\d+\.\d+"
       And content of file "newdir/newfile" should not match "192\.168\.\s+\.\d+"
       And content of file "newdir/newfile" is valid "IP"

   Scenario: Delete file
      When deleting file "newdir/newfile" succeeds
      Then file "newdir/newfile" should not exist

   Scenario: Delete directory
      When deleting directory "newdir" succeeds
      Then directory "newdir" should not exist

