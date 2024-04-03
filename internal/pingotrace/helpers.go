package pingotrace

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
)

func RemoveTrailingDot(s string) string {
	if len(s) > 0 && s[len(s)-1] == '.' {
		return s[:len(s)-1]
	}
	return s
}

func RemoveDuplicatesMap(keysMap map[string][]interface{}) map[string][]interface{} {
	cleanMap := make(map[string][]interface{})
	seen := make(map[string]bool)

	for key, val := range keysMap {
		if val[len(val)-1] == true { // Ensure that the last element is true
			// Handle case where key is IP
			ipInKey := net.ParseIP(key) != nil
			if ipInKey && !seen[key] {
				cleanMap[key] = val
				seen[key] = true
			}

			// Handle case where value is IP
			if ipInValue, ok := val[0].(string); ok {
				if net.ParseIP(ipInValue) != nil && !seen[ipInValue] {
					cleanMap[key] = val
					seen[ipInValue] = true
				}
			}
		} else { // If the last element is not true, just compare the keys
			if !seen[key] {
				cleanMap[key] = val
				seen[key] = true
			}
		}
	}
	return cleanMap
}

func RemoveDuplicatesList(elements []string) []string {
	// Use a map to record the existence of elements
	encountered := map[string]bool{}

	// Create a new empty slice
	result := []string{}

	// Iterate over the original slice
	for _, v := range elements {
		// Check if the element is recorded in the map
		if !encountered[v] && v != "Request timed out" {
			// If the element is not recorded, append it to the result slice
			// and record its existence in the map
			result = append(result, v)
			encountered[v] = true
		}
	}
	// Return the new slice which has no duplicates
	return result
}

func ReadFile(filepath string) string {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Failed opening file: %s", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Failed reading file: %s", err)
	}

	fileContent := strings.Join(lines, "\n")
	return fileContent
}
