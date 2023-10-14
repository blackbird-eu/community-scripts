package main

import (
	"os"
	"fmt"
	"bufio"
	"regexp"
	"net/url"
	"strings"
)

func FilterURLs(urls []string) []string {
	URLs := []string{}
	uniqueURLs := make(map[string]bool)

	for _, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
			continue
		}

		// Parse the URI path and query
		pathSegments := parsedURL.Path
		query := parsedURL.Query()

		// Ignore fragment and query parameters for comparison
		parsedURL.Fragment = ""
		parsedURL.RawQuery = ""
		
		// Match and replace UUIDs in the path with a placeholder like "00000000-0000-0000-0000-000000000000"
		re := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
		normalizedPath := re.ReplaceAllStringFunc(pathSegments, func(s string) string {
			return "00000000-0000-0000-0000-000000000000"
		})

		// Match and replace numerical path parameters with a placeholder like "1234"
		re = regexp.MustCompile(`/[0-9]+(/)?`)
		normalizedPath = re.ReplaceAllStringFunc(normalizedPath, func(s string) string {
			if strings.HasSuffix(s, "/") {
				return "/1234/"
			} else {
				return "/1234"
			}
		})

		// Update the path in the parsed URL
		parsedURL.Path = normalizedPath

		// Use the updated URL as the key for the map
		key := parsedURL.String()

		// Check if the key already exists
		if _, exists := uniqueURLs[key]; !exists {
			uniqueURLs[key] = true

			// Restore original values
			parsedURL.Path = pathSegments

			if len(query) > 0 {
				parsedURL, _ = url.Parse(fmt.Sprintf("%s?%s", key, query.Encode()))
			}

			URLs = append(URLs, parsedURL.String())
		}
	}

	return URLs
}

func main() {
	URLs := []string{}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		l := scanner.Text()
		URLs = append(URLs, fmt.Sprintf("%v", l))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Failed reading input from stdin!", err)
	}

	URLs = FilterURLs(URLs)

	// Print unique URLs
	for _, u := range URLs {
		fmt.Println(u)
	}
}
