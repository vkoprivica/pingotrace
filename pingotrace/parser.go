package pingotrace

import (
	"fmt"
	"regexp"
	"strings"
)

// Hostname + Hostname.domain + ipv4 parser without duplicates - subnet masks - wildcard mask
func ParseInput(text string) interface{} {
	// Define the regex pattern
	hostnameWithDomainsPattern := `^(?:https?:\/\/)?([a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*\.[a-zA-Z]{2,63})(?:\/|$)`
	hostnamePattern := `[a-zA-Z0-9-]{1,15}`
	subnetMaskPattern := `(0|255)\.(0|255)\.(0|255)\.(0|255)`

	// Compile the regex patterns
	hostnameWithDomainsRegex, err := regexp.Compile(hostnameWithDomainsPattern)
	if err != nil {
		return fmt.Sprintf("Error compiling regex: %s", err)
	}

	hostnameRegex, err := regexp.Compile(hostnamePattern)
	if err != nil {
		return fmt.Sprintf("Error compiling regex: %s", err)
	}

	subnetMaskRegex, err := regexp.Compile(subnetMaskPattern)
	if err != nil {
		return fmt.Sprintf("Error compiling subnet mask regex: %s", err)
	}

	// Initialize an empty slice to store unique matches
	uniqueMatchesSlice := make([]string, 0)
	// Initialize an empty map to keep track of added elements
	addedElements := make(map[string]bool)

	// Split the input by whitespace or newline
	substrings := regexp.MustCompile(`[\s\n]+`).Split(text, -1)

	// Iterate through the substrings to find matches
	for _, substring := range substrings {
		// Clean the substring of extra characters
		cleanedSubstring := strings.Trim(substring, `" ,`)

		// Check against subnet mask pattern, and continue if matched
		if subnetMaskRegex.MatchString(cleanedSubstring) {
			continue
		}

		// // Check against IPv4
		ipv4 := ParseIPv4(cleanedSubstring)
		if ipv4 != "" && !addedElements[ipv4] {
			uniqueMatchesSlice = append(uniqueMatchesSlice, ipv4)
			addedElements[ipv4] = true
			// You may or may not need "continue" here, depending on the rest of your logic.
			continue
		}

		// Check against hostname with domain pattern
		match := hostnameWithDomainsRegex.FindStringSubmatch(cleanedSubstring)
		if len(match) > 1 {
			if !addedElements[match[1]] {
				uniqueMatchesSlice = append(uniqueMatchesSlice, match[1])
				addedElements[match[1]] = true
				// fmt.Println(addedElements)
				continue
			}
		} else if hostnameRegex.MatchString(cleanedSubstring) && !addedElements[cleanedSubstring] {
			uniqueMatchesSlice = append(uniqueMatchesSlice, cleanedSubstring)
			addedElements[cleanedSubstring] = true
			// continue
		}
	}

	return uniqueMatchesSlice
}

func CheckHostnameWithDomain(text string) bool {
	hostnameWithDomainsPattern := `([a-zA-Z0-9-]+\.)+[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)?`
	matched, _ := regexp.MatchString(hostnameWithDomainsPattern, text)
	return matched
}

// CheckIPv4 checks if a string is a valid IPv4 address.
func CheckIPv4(ip string) bool {
	pattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	matched, _ := regexp.MatchString(pattern, ip)
	return matched
}

func ParseIPv4(text string) string {
	pattern := `((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`
	re := regexp.MustCompile(pattern)
	match := re.FindString(text)

	if len(match) > 0 {
		return match
	}

	return ""
}

// CheckIPv6 checks if a string is a valid IPv4 address.
func CheckIPv6(ip string) bool {
	pattern := `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`
	matched, _ := regexp.MatchString(pattern, ip)
	return matched
}

func ParseIPv6(text string) string {
	pattern := `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	re := regexp.MustCompile(pattern)
	match := re.FindString(text)

	if len(match) > 0 {
		return match
	}

	return ""
}
