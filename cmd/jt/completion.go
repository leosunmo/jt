package main

import (
	_ "embed"
	"fmt"
)

//go:embed completion/jt_completion.zsh
var completionZSH string

func printCompletionZSH() {
	fmt.Println(completionZSH)
}
