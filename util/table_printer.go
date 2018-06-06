//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package util

import (
	"fmt"
	"strconv"
	"strings"
)

// TableColumn objects can be printed with func PrintTable()
type TableColumn struct {
	Header string
	Data   []string
}

// PrintTable prints a list of TableColumns to the CLI with auto-sized columns and dividers
func PrintTable(tableColumns []TableColumn) {
	// Define table formatting
	baseFormatString := " %%-%ss  "
	dividerString := " | "

	// Vars for keeping track of column widths
	columnWidth := make(map[string]int)
	columnWidthStr := make(map[string]string)

	// Set appropriate column width for all columns
	for _, column := range tableColumns {
		for _, cellData := range column.Data {
			if len(cellData) > columnWidth[column.Header] {
				columnWidth[column.Header] = len(cellData)
				columnWidthStr[column.Header] = strconv.Itoa(len(cellData))
			}
		}
	}

	// Print column header
	for i, column := range tableColumns {
		formatString := fmt.Sprintf(baseFormatString, columnWidthStr[column.Header])
		if i < (len(tableColumns)) && i > 0 {
			fmt.Print(dividerString)
		}
		fmt.Printf(formatString, column.Header)
	}
	fmt.Println()

	// Print header to content divider (---)
	for i, column := range tableColumns {
		formatString := fmt.Sprintf(baseFormatString, columnWidthStr[column.Header])
		if i < (len(tableColumns)) && i > 0 {
			fmt.Print(dividerString)
		}
		fmt.Printf(formatString, strings.Repeat("-", columnWidth[column.Header]))
	}
	fmt.Println()

	// Print table contents
	for rowIndex := range tableColumns[0].Data {
		for i, column := range tableColumns {
			formatString := fmt.Sprintf(baseFormatString, columnWidthStr[column.Header])
			if i < (len(tableColumns)) && i > 0 {
				fmt.Print(dividerString)
			}
			fmt.Printf(formatString, column.Data[rowIndex])
		}
		fmt.Println()
	}
}
