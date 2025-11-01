// Copyright 2025 Alexandre Mahdhaoui
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"

	"github.com/alexandremahdhaoui/posix-yq/pkg/generator"
)

func main() {
	// Print shell script shebang
	fmt.Println(`#!/bin/sh
#
# Copyright 2025 Alexandre Mahdhaoui
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.`)
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
