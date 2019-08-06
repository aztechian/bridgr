package bridgr

import "log"

// Verbose determines whether debug logging is printed
var Verbose = false

// Println prints log messages based on the verbosity of the current instantiation of Bridgr
func Println(l ...interface{}) {
	log.Println(l...)
}

// Debugln works just like Println, however it will check the current setting for Verbosity
// and conditionally print the output if the user had asked for it.
func Debugln(l ...interface{}) {
	if Verbose {
		log.Println(l...)
	}
}

// Printf will print a formatted string to log output
func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Debugf behaves just like Printf, however it will conditionally output log messages only if the user
// has asked for verbose output.
func Debugf(format string, v ...interface{}) {
	if Verbose {
		log.Printf(format, v...)
	}
}

// Print prints the object(s) to log output
func Print(v ...interface{}) {
	log.Print(v...)
}

// Debug behaves just like Print, however it will conditionally output log messages only if the user has
// asked for verbose output.
func Debug(v ...interface{}) {
	if Verbose {
		log.Print(v...)
	}
}
