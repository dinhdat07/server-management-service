// Run: go run generate_fixtures.go
package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func main() {
	// 1. Valid: 3 servers
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", "server_name")
	f.SetCellValue("Sheet1", "B1", "ipv4")
	f.SetCellValue("Sheet1", "A2", "import-srv-01")
	f.SetCellValue("Sheet1", "B2", "10.0.0.1")
	f.SetCellValue("Sheet1", "A3", "import-srv-02")
	f.SetCellValue("Sheet1", "B3", "10.0.0.2")
	f.SetCellValue("Sheet1", "A4", "import-srv-03")
	f.SetCellValue("Sheet1", "B4", "10.0.0.3")
	f.SaveAs("servers.xlsx")
	fmt.Println("Created servers.xlsx")

	// 2. Empty: only headers
	f2 := excelize.NewFile()
	f2.SetCellValue("Sheet1", "A1", "server_name")
	f2.SetCellValue("Sheet1", "B1", "ipv4")
	f2.SaveAs("servers_empty.xlsx")
	fmt.Println("Created servers_empty.xlsx")

	// 3. Missing "ipv4" header
	f3 := excelize.NewFile()
	f3.SetCellValue("Sheet1", "A1", "server_name")
	f3.SetCellValue("Sheet1", "A2", "some-server")
	f3.SetCellValue("Sheet1", "B2", "10.0.0.10")
	f3.SaveAs("servers_no_header.xlsx")
	fmt.Println("Created servers_no_header.xlsx")
}
