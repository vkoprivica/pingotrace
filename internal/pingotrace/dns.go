package pingotrace

import (
	"context"
	"net"
	"strings"
	"sync"
)

// DNSLookup performs a DNS lookup on the provided host to get its IPv4 address.
// It returns the resolved IP address and a boolean indicating if the lookup was successful.
func DNSLookup(ctx context.Context, host string) (string, bool) {
	resultChan := make(chan string) // Channel to receive the resolved IP address
	errChan := make(chan error)     // Channel to receive any errors during resolution

	// Goroutine to resolve the host to an IP address
	go func() {
		ipAddr, err := net.ResolveIPAddr("ip4", host) // Resolve to IPv4 address
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- ipAddr.String()
	}()

	// Wait for a result, an error, or a cancellation signal
	select {
	case res := <-resultChan:
		return res, true
	case err := <-errChan:
		return err.Error(), false
	case <-ctx.Done():
		return "No DNS record found", false
	}
}

// PTRLookup performs a reverse DNS lookup (PTR) on the provided IPv4 address to get its associated domain name.
// It returns the associated domain name and a boolean indicating if the lookup was successful.
func PTRLookup(ctx context.Context, ipAddr string) (string, bool) {
	resultChan := make(chan string) // Channel to receive the domain name
	errChan := make(chan error)     // Channel to receive any errors during lookup

	// Goroutine to get the domain name associated with the IP address
	go func() {
		names, err := net.LookupAddr(ipAddr) // Perform reverse DNS lookup
		if err != nil || len(names) == 0 {
			errChan <- err
			return
		}
		// Trim trailing dot from the domain name
		resultChan <- strings.TrimSuffix(names[0], ".")
	}()

	// Wait for a result, an error, or a cancellation signal
	select {
	case res := <-resultChan:
		return res, true
	case <-errChan:
		return "PTR record lookup timed out", false
	case <-ctx.Done():
		return "PTR record lookup timed out", false
	}
}

// DNSPTR performs DNS and PTR lookups based on the inputs provided.
// If an input is an IP address, it will perform a PTR lookup. Otherwise, it does a DNS lookup.
// It returns a map containing the results and a slice of keys (inputs) in their original order.
func DNSPTR(ctx context.Context, inputs []string) (map[string][]interface{}, []string) {
	results := make(map[string][]interface{}) // Map to store the results
	keys := make([]string, 0, len(inputs))    // Slice to track the order of inputs

	var wg sync.WaitGroup // Synchronize goroutines
	// Struct for passing results between goroutines
	resultChan := make(chan struct {
		key     string
		address string
		success bool
	})

	for _, input := range inputs {
		// If the input is an IP address, perform PTR lookup
		if CheckIPv4(input) {
			wg.Add(1)
			go func(input string) {
				defer wg.Done()
				domainName, success := PTRLookup(ctx, input)
				resultChan <- struct {
					key     string
					address string
					success bool
				}{input, domainName, success}
			}(input)
		} else { // Otherwise, perform DNS lookup
			wg.Add(1)
			go func(input string) {
				defer wg.Done()
				ipAddr, success := DNSLookup(ctx, input)
				resultChan <- struct {
					key     string
					address string
					success bool
				}{input, ipAddr, success}
			}(input)
		}
		keys = append(keys, input) // Track the order of the inputs
	}

	// Goroutine to close the results channel once all lookups are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results from the channel
	for res := range resultChan {
		results[res.key] = []interface{}{res.address, res.success}
	}

	return results, keys
}

// DNSPTR to IP performs DNS and PTR lookups based on the inputs provided.
// If an input is an IP address, it will perform a PTR lookup. Otherwise, it does a DNS lookup.
// It returns a map containing the results and a slice of keys (inputs) in their original order.
func DNSPTRtoIP(ctx context.Context, inputs []string) []string {
	type result struct {
		index   int
		address string
		success bool
	}

	results := make([]result, len(inputs)) // Use a slice to store results in order
	var wg sync.WaitGroup
	resultChan := make(chan result, len(inputs)) // Buffered channel

	for i, input := range inputs {
		wg.Add(1)
		go func(i int, input string) {
			defer wg.Done()
			var res result
			res.index = i // Capture the index
			if CheckIPv4(input) {
				res.address = input
				res.success = true
			} else {
				res.address, res.success = DNSLookup(ctx, input)
			}
			resultChan <- res
		}(i, input) // Pass the current index and input
	}

	// Goroutine to close the results channel once all lookups are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for res := range resultChan {
		results[res.index] = res // Place the result in its original position
	}

	// Construct the final slice of IP addresses, maintaining the original order
	var ipAddresses []string
	for _, res := range results {
		if res.success {
			ipAddresses = append(ipAddresses, res.address)
		} else {
			// Handle failed lookups if necessary
			ipAddresses = append(ipAddresses, "Lookup failed")
		}
	}

	return ipAddresses
}
