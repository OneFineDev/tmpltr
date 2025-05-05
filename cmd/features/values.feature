Feature: get values
    A user should be able to get a yaml representation of the values required in a SourceSet or Sources list

Scenario Outline: getting the values for a SourceSet
    Given a valid SourceSet in the source.yaml file
    When a user runs tmpltr get values --sourceSet "<sourceSet>"
    Then the desired values should be printed to the console
