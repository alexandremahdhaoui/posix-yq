package main

import (
	"fmt"
)

func main() {
	// Print shell script shebang
	fmt.Println("#!/bin/sh")
	fmt.Println()

	// Concatenate all generator modules in the correct order
	fmt.Print(GenerateShellHeader())
	fmt.Println()
	
	fmt.Print(GenerateParser())
	fmt.Println()
	
	fmt.Print(GenerateCoreFunctions())
	fmt.Println()
	
	fmt.Print(GenerateAdvancedFunctions())
	fmt.Println()
	
	fmt.Print(GenerateOperators())
	fmt.Println()
	
	fmt.Print(GenerateJSON())
	fmt.Println()
	
	fmt.Print(GenerateEntryPoint())
}
