package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var wg sync.WaitGroup

func main() {

	clearScreen()
	printLogo()

	domainPtr := flag.String("d", "", "Domain to check (e.g., https://google.com/)")
	domainLongPtr := flag.String("domain", "", "Domain to check (e.g., https://google.com/)")
	wordlistPtr := flag.String("w", "", "Path to the wordlist file")
	wordlistLongPtr := flag.String("wordlist", "", "Path to the wordlist file")
	flag.Parse()

	if *domainPtr == "" && *domainLongPtr == "" || *wordlistPtr == "" && *wordlistLongPtr == "" {
		fmt.Println("Usage:   subsearch -d|--domain <domain> -w|--wordlist <wordlist>")
		fmt.Println("Example: subsearch --domain http/s://example.com --wordlist sub.txt\n")
		os.Exit(1)
	}

	domain := *domainPtr
	if *domainLongPtr != "" {
		domain = *domainLongPtr
	}

	wordlist := *wordlistPtr
	if *wordlistLongPtr != "" {
		wordlist = *wordlistLongPtr
	}

	wordlistFile, err := os.Open(wordlist)
	if err != nil {
		fmt.Println("Error opening wordlist:", err)
		os.Exit(1)
	}
	defer wordlistFile.Close()

	scanner := bufio.NewScanner(wordlistFile)
	for scanner.Scan() {
		subdomain := scanner.Text()
		wg.Add(1)
		go checkSubdomain(domain, subdomain)
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading wordlist:", err)
		os.Exit(1)
	}
}

func checkSubdomain(domain, subdomain string) {
	defer wg.Done()
	url := constructURL(domain, subdomain)
	statusCode := getRequestStatusCode(url)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	statusColor := color.New(color.Bold, color.FgHiRed).SprintFunc()
	boldFont := color.New(color.Bold).SprintFunc()
	if statusCode == http.StatusOK {
		fmt.Printf("%s  %s  %s\n", boldFont(timestamp), statusColor("VALID"), boldFont(url))
	}
}

func printLogo() {
	logo := ` _____ _____ _____ _____ _____ _____ _____ _____ _____ {1.0.1}
|   __|  |  | __  |   __|   __|  _  | __  |     |  |  |
|__   |  |  | __ -|__   |   __|     |    -|   --|     |
|_____|_____|_____|_____|_____|__|__|__|__|_____|__|__|
`
	fmt.Println(color.New(color.Bold).Sprint(logo))
}

func clearScreen() {
	cmd := exec.Command(clearCommand())
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func clearCommand() string {
	if runtime.GOOS == "windows" {
		return "cls"
	}
	return "clear"
}

func constructURL(domain, subdomain string) string {
	if strings.HasPrefix(domain, "https://") || strings.HasPrefix(domain, "http://") {
		return fmt.Sprintf("%s://%s.%s", getScheme(domain), subdomain, getDomainName(domain))
	}
	return fmt.Sprintf("https://%s.%s%s", subdomain, getDomainName(domain), domain)
}

func getDomainName(url string) string {
	parts := strings.Split(url, "//")
	domain := strings.TrimSuffix(parts[1], "/")
	return domain
}

func getScheme(url string) string {
	if strings.HasPrefix(url, "https://") {
		return "https"
	}
	return "http"
}

func getRequestStatusCode(url string) int {
	response, err := http.Head(url)
	if err != nil {
		return -1
	}
	return response.StatusCode
}
