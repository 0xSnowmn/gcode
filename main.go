package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/gookit/color"
	"github.com/jessevdk/go-flags"
	"github.com/mbndr/figlet4go"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Setting Options
var opts struct {
	// File For urls
	Filename string `short:"f" long:"file" description:"subdomains file" required:"true"`
	// Timeout For Request
	Timeout int `short:"t" long:"timeout" description:"Timeout For Requests" `
	// Concurrency For the requests
	Concurrency int `short:"c" long:"concurrency" default:"25" description:"Concurrency For Requests"`
}

// Ouptut Colors
var (
	errColor  = color.Style{color.BgRed, color.OpBold}.Render
	WarnColor = color.Style{color.FgYellow, color.OpBold}.Render
	SuColor   = color.Style{color.Green, color.OpBold}.Render
	errHost   = color.Style{color.FgLightWhite, color.OpBold}.Render
	reColor   = color.Style{color.Blue, color.OpBold}.Render
	fooColor  = color.Style{color.Red, color.OpBold}.Render
)

func main() {
	// Parse Args
	_, err := flags.Parse(&opts)

	// Check for any error occurred
	if err != nil {
		fmt.Println(errColor("try to use -h to show the options and usage :)"))
		// Exit from script after print this
		return
	}
	// setting vars for options
	file := opts.Filename
	timeout := time.Duration(opts.Timeout * 1000000)
	conc := opts.Concurrency
	// Check if file exists or not
	if !isExists(file) {
		fmt.Println(errColor("File not found!"))
		// Exit from script after print this
		return
	}
	// Print the banner
	banner()
	fmt.Println("")
	// Get urls from chan (geturls func)
	urls := geturls(file)
	// Set WaitGroup
	var wg sync.WaitGroup
	// Looping the Concurrency
	for i := 0; i < conc; i++ {
		// everytime looping increase one for WaitGroup
		wg.Add(1)
		// Start Go Routine
		go func() {
			// Done the sync
			defer wg.Done()
			// Start urls chan loop
			for url := range urls {
				// Get Status Code For every url in chan
				status(url, timeout)
			}
		}()
	}
	// Wait untill workers finish
	wg.Wait()
}

// Check For file exists func
func isExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

// Get Status Func
func status(url string, timeout time.Duration) {
	// Add sSa
	url = "https://" + url
	// Editing http transport
	trans := &http.Transport{
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
		// Skip Certificate Error
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 5 * time.Second,
		Dial: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: time.Second,
		}).Dial,
	}
	// Editing http client
	client := &http.Client{
		// Passing transport var
		Transport: trans,
		// prevent follow redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout * time.Second,
	}
	// Start an new HEAD Requets
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return
	}
	// Close Connection
	req.Header.Set("Connection", "close")
	// Do the request
	resp, err := client.Do(req)
	if err != nil {
		// Check if host is up or not
		fmt.Println(errHost(url) + " -> " + errHost("No Such Host"))
		return
	}
	// Close the response body
	defer resp.Body.Close()
	// Convert status_code from int to str
	stat := strconv.Itoa(resp.StatusCode)
	// Check for the first char(3xx) for the status code and print the location header
	if stat[0:1] == "3" {
		fmt.Println(reColor(url) + " : " + reColor(stat) + " -> " + WarnColor(resp.Header.Get("Location")))
	} else if stat[0:1] == "4" {
		df := url + " -> " + stat
		fmt.Println(fooColor(df))
	} else if stat[0:1] == "2" {
		fmt.Println(SuColor(url + " -> " + stat))
	}
}

// Get urls from the file

func geturls(filename string) <-chan string {
	// Create the channel to store url into it
	urls := make(chan string)

	file, _ := os.Open(filename)
	// Create a new Scanner
	sc := bufio.NewScanner(file)
	// Start Go Routine
	go func() {
		// Close the file
		defer file.Close()
		// Close the urls Channel
		defer close(urls)
		// Looping lines and store all urls into urls channel
		for sc.Scan() {
			// Convert all urls to lower
			subdomain := strings.ToLower(sc.Text())
			// Store subdomain into urls
			urls <- subdomain
		}
	}()

	return urls
}

func banner() {
	ascii := figlet4go.NewAsciiRender()
	options := figlet4go.NewRenderOptions()
	// Colors For Banner Output
	options.FontColor = []figlet4go.Color{
		figlet4go.ColorGreen,
		figlet4go.ColorYellow,
		figlet4go.ColorCyan,
	}

	renderStr, _ := ascii.RenderOpts("GCODE", options)
	fmt.Print(renderStr)
	we := center("v1, author: @yghonem14")
	fmt.Println(we)
}

// Print string in center of terminal

func center(str string) string {
	ss, _ := fmt.Printf(fmt.Sprintf("%%-%ds", 90/2), fmt.Sprintf(fmt.Sprintf("%%%ds", 80/2), str))
	text := string(ss)
	// Reomve "(" from the end of string
	return strings.Trim(text, "(")
}
