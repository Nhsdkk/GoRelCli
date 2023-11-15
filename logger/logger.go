package logger

import (
	"fmt"
	"time"
)

func LogStep(stepName string, function func() error) error {
	fmt.Println(fmt.Sprintf("Starting task with name '%s'...", stepName))
	startTime := time.Now()
	if err := function(); err != nil {
		fmt.Println(fmt.Sprintf("Task with name '%s' failed in %v ms with error:\n %s", stepName, time.Since(startTime).Milliseconds(), err))
		return err
	}
	fmt.Println(fmt.Sprintf("Task with name '%s' successfully ended in %v ms", stepName, time.Since(startTime).Milliseconds()))
	return nil
}
