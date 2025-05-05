<!--  GO -->

<!-- Testing -->

When generating unit tests for go, use table-driven tests where appropriate and always use the arrange, act, assert pattern and indicate these
"regions" with comments, even if there is nothing to do in that region.

When testing the parsing or manipulation of config files (yaml, json, toml), always embed a literal document in the test case itself. Rather than being arbitrary values, they should be like configurations as they would be in normal use of the application.

When testing unexported methods or functions, these tests should be created in a file called `export_test.go` in the same package as the tested method or function.

Wherever the tested method or function uses the afero package, always use a memmap fs as the filesystem in tests (i.e. don't write to disk where this can be avoided.).

When writing assertions in tests, do not abstract assertion/verification logic into functions stores in test cases, explicitly define the assertion or verification logic in the "Assert" portion of the test; use conditional logic if needed here. Use testify require (not testify assert) when checking for the presence or absence of errors. Use assert.Len when asserting the length of objects.

When creating a test file, the package name should always be the package name with "\_test" appended.

When creating or satisfying a slog logger on services, always use the DiscardHandler.

<!-- Style -->

Any comments you include above functions or types should end with a period.

<!-- Code Quality -->

When generating go code, avoid variable name shadowing.
