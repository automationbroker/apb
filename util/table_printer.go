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

// TableSettings provides an optional way to change default PrintTable() settings
type TableSettings struct {
	BaseFormatString string
	DividerString    string
}

// Default table formatting
const defaultBaseFormatString = " %%-%ss  "
const defaultDividerString = " | "

// PrintTable prints a list of TableColumns with auto-sized columns and dividers.
func PrintTable(columns []*TableColumn, settings *TableSettings) {
	// Define table formatting
	var baseFormatString string
	var dividerString string

	if settings == nil {
		baseFormatString = defaultBaseFormatString
		dividerString = defaultDividerString
	} else {
		baseFormatString = settings.BaseFormatString
		dividerString = settings.DividerString
	}

	// Vars for keeping track of column widths
	columnWidth := make(map[string]int)
	columnWidthStr := make(map[string]string)

	// Set appropriate column width for all columns
	for _, column := range columns {
		for _, cellData := range column.Data {
			if len(cellData) > columnWidth[column.Header] {
				columnWidth[column.Header] = len(cellData)
				columnWidthStr[column.Header] = strconv.Itoa(len(cellData))
			}
		}
	}

	// Print column header
	for i, column := range columns {
		formatString := fmt.Sprintf(baseFormatString, columnWidthStr[column.Header])
		if i < (len(columns)) && i > 0 {
			fmt.Print(dividerString)
		}
		fmt.Printf(formatString, column.Header)
	}
	fmt.Println()

	// Print header to content divider (---)
	for i, column := range columns {
		formatString := fmt.Sprintf(baseFormatString, columnWidthStr[column.Header])
		if i < (len(columns)) && i > 0 {
			fmt.Print(dividerString)
		}
		fmt.Printf(formatString, strings.Repeat("-", columnWidth[column.Header]))
	}
	fmt.Println()

	// Print table contents
	for rowIndex := range columns[0].Data {
		for i, column := range columns {
			formatString := fmt.Sprintf(baseFormatString, columnWidthStr[column.Header])
			if i < (len(columns)) && i > 0 {
				fmt.Print(dividerString)
			}
			fmt.Printf(formatString, column.Data[rowIndex])
		}
		fmt.Println()
	}
}
