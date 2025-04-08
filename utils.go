package main

import "fmt"

func format(template string, a ...any) string {
	return fmt.Sprintf(template, a...)
}

func exception(err error, propmt string) bool {
	if err != nil {
		fmt.Sprintln(propmt+": ", err)
		return true
	}
	return false
}
