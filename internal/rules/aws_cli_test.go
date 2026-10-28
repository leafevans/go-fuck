package rules

import "testing"

// awsCLINoSuggestions is the aws output without a suggestion.
const awsCLINoSuggestions = `usage: aws [options] <command> <subcommand> [<subcommand> ...] [parameters]
To see help text, you can run:

  aws help
  aws <command> help
  aws <command> <subcommand> help
aws: error: argument command: Invalid choice, valid choices are:

dynamodb                                 | dynamodbstreams
ec2                                      | ecr
`

// awsCLIMisspelledCommand is the aws output for a misspelled command.
const awsCLIMisspelledCommand = `usage: aws [options] <command> <subcommand> [<subcommand> ...] [parameters]
To see help text, you can run:

  aws help
  aws <command> help
  aws <command> <subcommand> help
aws: error: argument command: Invalid choice, valid choices are:

dynamodb                                 | dynamodbstreams
ec2                                      | ecr


Invalid choice: 'dynamdb', maybe you meant:

  * dynamodb
`

// awsCLIMisspelledSubcommand is the aws output for a misspelled subcommand.
const awsCLIMisspelledSubcommand = `usage: aws [options] <command> <subcommand> [<subcommand> ...] [parameters]
To see help text, you can run:

  aws help
  aws <command> help
  aws <command> <subcommand> help
aws: error: argument operation: Invalid choice, valid choices are:

query                                    | scan
update-item                              | update-table


Invalid choice: 'scn', maybe you meant:

  * scan
`

// awsCLIMisspelledSubcommandWithMultipleOptions is the aws output with two
// suggestions.
const awsCLIMisspelledSubcommandWithMultipleOptions = `usage: aws [options] <command> <subcommand> [<subcommand> ...] [parameters]
To see help text, you can run:

  aws help
  aws <command> help
  aws <command> <subcommand> help
aws: error: argument operation: Invalid choice, valid choices are:

describe-table                           | get-item
list-tables                              | put-item


Invalid choice: 't-item', maybe you meant:

  * put-item
  * get-item
`

func TestRuleAwsCLI(t *testing.T) {
	resetTestState(t)

	rule := Get("aws_cli")
	if rule == nil {
		t.Fatal("aws_cli rule is not registered")
	}

	assertMatch(t, rule, "aws dynamdb scan", awsCLIMisspelledCommand)
	assertMatch(t, rule, "aws dynamodb scn", awsCLIMisspelledSubcommand)
	assertMatch(t, rule, "aws dynamodb t-item", awsCLIMisspelledSubcommandWithMultipleOptions)

	assertNotMatch(t, rule, "aws dynamodb invalid", awsCLINoSuggestions)

	assertNewCommands(t, rule, "aws dynamdb scan", awsCLIMisspelledCommand,
		[]string{"aws dynamodb scan"})
	assertNewCommands(t, rule, "aws dynamodb scn", awsCLIMisspelledSubcommand,
		[]string{"aws dynamodb scan"})
	assertNewCommands(t, rule, "aws dynamodb t-item", awsCLIMisspelledSubcommandWithMultipleOptions,
		[]string{"aws dynamodb put-item", "aws dynamodb get-item"})
}
