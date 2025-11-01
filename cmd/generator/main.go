package main

import (
	"fmt"

	"github.com/alexandremahdhaoui/posix-yq/pkg/generator"
)

func main() {
	// Print shell script shebang
	fmt.Println("#!/bin/sh")
	fmt.Println()

	// Concatenate all generator modules in the correct order
	fmt.Print(generator.GenerateShellHeader())
	fmt.Println()

	fmt.Print(generator.GenerateParser())
	fmt.Println()

	fmt.Print(generator.GenerateCoreFunctions())
	fmt.Println()

	fmt.Print(generator.GenerateAdvancedFunctions())
	fmt.Println()

	fmt.Print(generator.GenerateOperators())
	fmt.Println()

	fmt.Print(generator.GenerateJSON())
	fmt.Println()

	fmt.Print(generator.GenerateEntryPoint())
}
