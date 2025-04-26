Feature: create project
    Users can create projects with a range of configuration options and sources

Scenario Outline: using SourceSets (defined in a local .sources.yaml) using different authentication methods
    Given a sourceSet of type "<sourceSet>" and an authentication method of "<authType>"
    When a user runs the create project command with the flags "<flagString>" and the provided "<valuesFile>"
    Then desired files are copied to target path

    Examples:
        | sourceSet          | authType | flagString | valuesFile  |
        | terraformChildSet  | ssh      | --token    | values.yaml |


Scenario Outline: title
    Given context
    When event
    Then outcome
