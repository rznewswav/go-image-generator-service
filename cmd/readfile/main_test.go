package main

import (
	"testing"
)

func TestReadImage(t *testing.T) {
	Run(weeklyTemplate, "/tmp/out.png", M{
		"{weekly total cases}": "9,548 cases that is very l",
	})
}
