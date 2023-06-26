package main

import (
	"fmt"
	"flag"
	"regexp"
	"strings"
	"net/http"
	"io/ioutil"
)

func main() {
	targetFlag := flag.String("target", "", "Specify your target domain name")
	flag.Parse()

	if *targetFlag == "" {
		flag.Usage()
		return
	}

	var target string = *targetFlag

	tag := FetchGTMTag(target)
	subdomains := FetchDomains(target, tag)

	for _, s := range subdomains {
		fmt.Println(s)
	}
}

func FetchGTMTag(target string) string {
	var tag string = ""

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s", target), nil)
	if err != nil {
		fmt.Printf("ERROR: Failed to send request %s (%s)\n", req, err)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR: Failed to read response %s (%s)\n", res, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return tag
	}

	// Match and return a GTM ID if present
	re := regexp.MustCompile(`GTM-[A-Z0-9]{7}`)
	tag = re.FindString(string(body))

	fmt.Println("Tag:", tag)

	return tag
}

func FetchDomains(target string, tag string) []string {
	var subdomains []string

	req, err := http.NewRequest("GET", fmt.Sprintf("https://googletagmanager.com/gtm.js?id=%s", tag), nil)
	if err != nil {
		fmt.Printf("ERROR: Failed to send request %s (%s)\n", req, err)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR: Failed to read response %s (%s)\n", res, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return subdomains
	}

	// Root domain used to match domains
	rootDomain := ParseRootDomain(target)

	fmt.Println("rootDomain:", rootDomain)

	re := regexp.MustCompile(fmt.Sprintf(`(([a-zA-Z0-9-\.]+)?\.)?%s\.[a-zA-Z]{0,3}(\.[a-zA-Z]{0,3})?`, rootDomain))
	domains := re.FindAllString(string(body), -1)

	for _, domain := range domains {
		// Inscope regexp to match inscope domains
		inScope, _ := regexp.MatchString(fmt.Sprintf(`^(.*\.)?%s$`, target), domain)
		
		if inScope {
			subdomains = append(subdomains, domain)
		}
	}

	return subdomains
}

func ParseRootDomain(target string) string {
	// Remove any subdomains
	domainParts := strings.Split(target, ".")
	numParts := len(domainParts)

	var rootDomain string = ""

	// Check if the TLD is a two-letter country code or a single part TLD
	if numParts >= 3 && (len(domainParts[numParts-1]) == 2 || len(domainParts[numParts-1]) == 3) {
		rootDomain = domainParts[numParts-3]
	} else if numParts >= 2 {
		rootDomain = domainParts[numParts-2]
	}

	return rootDomain
}
