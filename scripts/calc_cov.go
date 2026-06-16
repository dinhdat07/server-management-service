package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
)

func main() {
	file, err := os.Open("coverage.out")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	var totalStmts int64
	var coveredStmts int64
	
	pkgCov := make(map[string]int64)
	pkgTotal := make(map[string]int64)

	// Skip the first line: "mode: set"
	if scanner.Scan() {
		// Do nothing
	}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) < 3 {
			continue
		}
		
		path := parts[0]
		
		// Filter criteria: we ONLY want internal/modules/
		if !strings.HasPrefix(path, "server-management-service/internal/modules/") {
			continue
		}
		if strings.Contains(path, "/mock/") || strings.Contains(path, "/domain/") || strings.Contains(path, "/gen/") {
			continue
		}

		// Parts[1] is stmts, Parts[2] is count
		stmts, _ := strconv.ParseInt(parts[1], 10, 64)
		count, _ := strconv.ParseInt(parts[2], 10, 64)

		// Path looks like: server-management-service/internal/modules/identity/service
		// Split by "/"
		partsPath := strings.Split(path, "/")
		var moduleName string
		for i, p := range partsPath {
			if p == "modules" && i+1 < len(partsPath) {
				moduleName = partsPath[i+1]
				break
			}
		}
		if moduleName == "" {
			continue
		}

		totalStmts += stmts
		pkgTotal[moduleName] += stmts
		if count > 0 {
			coveredStmts += stmts
			pkgCov[moduleName] += stmts
		}
		
		// calc line coverage
		if stmts > 0 {
		    _ = float64(count) / float64(stmts) // wait count is execution count, not covered statement.
		    // if count > 0, it means the block was executed. The file has block level granularity.
		    // let's not calculate per file here, it's easier to just run the command.
		}
	}

	if totalStmts == 0 {
		fmt.Println("No matching statements found")
		return
	}

	for pkg, cov := range pkgCov {
		fmt.Printf("%s: %.2f%%\n", pkg, float64(cov)/float64(pkgTotal[pkg])*100)
	}

	percentage := float64(coveredStmts) / float64(totalStmts) * 100
	fmt.Printf("Core Modules (Testable) Coverage: %.2f%%\n", percentage)
	fmt.Printf("Covered Statements: %d\n", coveredStmts)
	fmt.Printf("Total Statements: %d\n", totalStmts)
}
