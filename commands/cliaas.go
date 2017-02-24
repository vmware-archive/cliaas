package commands

type CliaasCommand struct {
	Version VersionCommand `command:"version" description:"Print version information and exit"`
	AWS     AWSCommand     `command:"aws" description:"blue-green deployment of aws VM matching name pattern"`
}

//Cliaas - the command structure
var Cliaas CliaasCommand
