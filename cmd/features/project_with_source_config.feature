Feature: create project
    Users can create a project taking advantage of all elements of the source.yaml file, including:
    - sourceSets
    - sources
    - sourceAuth
    - sourceValues

Background:

# Scenario Outline: using a simple source list of sources without authentication, without values
#     Given source file "<sourcesFile>"
#     When a user runs tmpltr create project
#     Then desired files are copied to target path

#     Examples:
#         | sourcesFile |
#         | gitSources.yaml |
#         | blobSources.yaml |
#         | mixedSources.yaml |

# Scenario Outline: using a simple source list of sources with authentication
#     Given a source list type of "<sourceType>" and an authentication method of "<authType>"
#     When a user runs tmpltr create project --sourceFile "<sourcesFile>" --targetPath ./output
#     Then a the desired files should be copied to the project and any values replaced

#     Examples:
#         | sourceType | sourcesFile | authType |
#         | git | gitSources.yaml | pat via env var |
#         | git | gitSources.yaml | pat via sourceAuth var |
#         | git | gitSources.yaml | ssh  |
#         | blob | blobSources.yaml | key |
#         | mixed | mixedSources.yaml | key and pat |
#         | mixed | mixedSources.yaml | key and ssh |
  

Scenario Outline: using sourceSets defined in ~/.tmpltr/.sources.yaml
    Given a sourceSet of type "<sourceSet>"
    When a user runs tmpltr create project --sourceSet "<sourceSet>" --targetPath ./output
    Then the desired "<sourceSet>" files should be copied to the project folder and any values replaced

    Examples:
        | sourceSet | 
        | terraformChild  |
        | terraformDeployment |
        | goWeb |

# Scenario Outline: checking cached versions of sources in sourceSets
#     Given --cacheSources flag is set to "<cacheSourcesFlag>" and .tmpltr cacheSources setting is "<cacheSourcesSetting>" 
#     When a user runs tmpltr create project
#     Then a the user "<shouldOrShouldNot>" be prompted to update the source

#     Examples:
#         | cacheSourcesFlag | cacheSourcesSetting | shouldOrShouldNot |
#         | true | true | true |
#         | true | false | true |
#         | false | true | false |
#         | false | false | false |