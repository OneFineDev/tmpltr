<!--  GO -->

<!-- Testing -->

When generating unit tests for go, use table-driven tests where appropriate and always use the arrange, act, assert pattern and indicate these "regions" with comments, even if there is nothing to do in that region.

Wherever the tested method or function uses the afero package, always use a memmap fs as the filesystem in tests (i.e. don't write to disk where this can be avoided.).

When writing assertions in tests, do not abstract assertion/verification logic into functions stores in test cases, explicitly define the assertion or verification logic in the "Assert" portion of the test; use conditional logic if needed here.

When creating a test file, the package name should always be the package name with "\_test" appended.

<!-- Style -->

Any comments you include above functions or types should end with a period.
