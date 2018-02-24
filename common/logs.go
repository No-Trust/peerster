package common

/*
   Imports
*/

import "fmt"

/*
   Constants
*/

/**
 * Constants defining the various log modes for the program
 */
const (

	// Output no logs
	LOG_MODE_NONE string = "none"

	// Only log events such as rumors received
	LOG_MODE_REACTIVE string = "reactive"

	// Log everything
	LOG_MODE_FULL string = "full"
)

/*
   Variables
*/

/**
 * A map describing relationship between different log modes.
 * Each mode M is mapped to a list of other modes that also
 * need to be accepted if M is an accepted mode.
 * For example, "A -> nil" implies that if mode A is accepted,
 * then no additional modes are to be accepted, whereas "A ->
 * { B , C } implies that if mode A is accepted, then modes
 * B and C are also to be accepted, and their entries in the
 * map need to be checked recursively in order to add potential
 * modes related to them.
 */
var logModeGraph map[string][]string = map[string][]string{

	LOG_MODE_NONE: nil,

	LOG_MODE_REACTIVE: nil,

	LOG_MODE_FULL: {LOG_MODE_REACTIVE},
}

/**
 * A map holding as keys the log modes that are to be accepted
 * during the current execution, that is the modes that, when
 * passed as target modes to the Log function, should result in
 * a message being logged.
 * The values of the map are completely irrelevant in this case
 * (as the map is used as a convenient replacement of a set, to
 * avoid having duplicates) and will always be given the boolean
 * value `true`.
 */
var acceptedLogModes map[string]bool = map[string]bool{LOG_MODE_NONE: true}

/*
   Functions
*/

/**
 * Recursively traverses the log mode graph starting with the
 * given mode in order to determine the accepted log modes for
 * the current execution.
 */
func InitLogger(mode string) {

	delete(acceptedLogModes, LOG_MODE_NONE)

	modes := []string{mode}

	for len(modes) > 0 {

		acceptedLogModes[modes[0]] = true

		modes = append(modes[1:], logModeGraph[modes[0]]...)

	}

}

/**
 * Logs the given message if the given target mode is an
 * accepted log mode.
 */
func Log(message, targetMode string) {

	if _, ok := acceptedLogModes[LOG_MODE_NONE]; ok {
		return
	}

	if _, ok := acceptedLogModes[targetMode]; ok {
		fmt.Println(message)
	}

}
