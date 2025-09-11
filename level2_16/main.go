package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// Downloader represents the website downloader.
type Downloader struct {
	baseURL    *url.URL
	dir        string
	visited    map[string]bool
	mu         sync.Mutex
	wg         sync.WaitGroup
	maxDepth   int
	concurrent int
	sem        chan struct{}
	robots     *RobotsTxt
	client     *http.Client
}

// RobotsTxt holds parsed robots.txt rules.
type RobotsTxt struct {
	disallows map[string][]string
}

// NewDownloader creates a new Downloader instance.
func NewDownloader(baseURLStr string, dir string, maxDepth int, concurrent int) (*Downloader, error) {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}

	d := &Downloader{
		baseURL:    baseURL,
		dir:        dir,
		visited:    make(map[string]bool),
		maxDepth:   maxDepth,
		concurrent: concurrent,
		sem:        make(chan struct{}, concurrent),
		robots:     &RobotsTxt{disallows: make(map[string][]string)},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Load robots.txt
	if err := d.loadRobotsTxt(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load robots.txt: %v\n", err)
	}

	return d, nil
}

// loadRobotsTxt downloads and parses robots.txt.
func (d *Downloader) loadRobotsTxt() error {
	robotsURL := d.baseURL.Scheme + "://" + d.baseURL.Host + "/robots.txt"
	resp, err := d.client.Get(robotsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("robots.txt returned status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	currentUserAgent := "*"
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		switch parts[0] {
		case "User-agent:":
			currentUserAgent = parts[1]
		case "Disallow:":
			d.robots.disallows[currentUserAgent] = append(d.robots.disallows[currentUserAgent], parts[1])
		}
	}
	return scanner.Err()
}

// isAllowedByRobots checks if the URL is allowed by robots.txt.
func (d *Downloader) isAllowedByRobots(u string) bool {
	parsedU, err := url.Parse(u)
	if err != nil {
		return false
	}
	path := parsedU.Path
	if path == "" {
		path = "/"
	}

	for _, disallow := range d.robots.disallows["*"] {
		if strings.HasPrefix(path, disallow) {
			return false
		}
	}
	return true
}

// Download starts the website downloading process.
func (d *Downloader) Download() error {
	d.wg.Add(1)
	d.download(d.baseURL.String(), 0)
	d.wg.Wait()
	return nil
}

// download downloads a page or resource.
func (d *Downloader) download(u string, depth int) {
	defer d.wg.Done()

	if depth > d.maxDepth {
		return
	}

	if !d.isAllowedByRobots(u) {
		fmt.Fprintf(os.Stderr, "URL %s disallowed by robots.txt\n", u)
		return
	}

	d.mu.Lock()
	if d.visited[u] {
		d.mu.Unlock()
		return
	}
	d.visited[u] = true
	d.mu.Unlock()

	d.sem <- struct{}{}
	fmt.Printf("Starting download: %s (depth %d, active %d/%d)\n", u, depth, len(d.sem), d.concurrent)
	defer func() {
		<-d.sem
		fmt.Printf("Finished download: %s (active %d/%d)\n", u, len(d.sem), d.concurrent)
	}()

	resp, err := d.client.Get(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading %s: %v\n", u, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "error: %s returned status %d\n", u, resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading body: %v\n", err)
		return
	}

	localPath, err := d.urlToLocalPath(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting local path: %v\n", err)
		return
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating directory: %v\n", err)
		return
	}

	// Rewrite HTML to use local paths
	if strings.HasSuffix(localPath, ".html") || strings.HasSuffix(localPath, "/") {
		rewritten := d.rewriteHTML(string(body), u)
		if err := os.WriteFile(localPath, []byte(rewritten), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error saving file: %v\n", err)
			return
		}
		d.parseHTML(string(body), u, depth)
	} else {
		if err := os.WriteFile(localPath, body, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error saving file: %v\n", err)
			return
		}
	}
}

// rewriteHTML rewrites URLs in HTML to local paths.
func (d *Downloader) rewriteHTML(htmlContent string, base string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing HTML for rewrite: %v\n", err)
		return htmlContent
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i := range n.Attr {
				if n.Attr[i].Key == "href" || n.Attr[i].Key == "src" {
					fullURL := d.resolveURL(n.Attr[i].Val, base)
					if fullURL != "" {
						localPath, err := d.urlToLocalPath(fullURL)
						if err == nil {
							relPath, err := filepath.Rel(filepath.Dir(localPath), localPath)
							if err == nil {
								n.Attr[i].Val = relPath
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var b strings.Builder
	if err := html.Render(&b, doc); err != nil {
		fmt.Fprintf(os.Stderr, "error rendering HTML: %v\n", err)
		return htmlContent
	}
	return b.String()
}

// parseHTML parses the HTML content and downloads resources and links.
func (d *Downloader) parseHTML(htmlContent string, base string, depth int) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing HTML: %v\n", err)
		return
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				for i := 0; i < len(n.Attr); i++ {
					if n.Attr[i].Key == "href" {
						link := n.Attr[i].Val
						fullLink := d.resolveURL(link, base)
						if fullLink != "" && d.isSameDomain(fullLink) {
							d.wg.Add(1)
							go d.download(fullLink, depth+1)
						}
					}
				}
			case "img", "link", "script":
				for i := 0; i < len(n.Attr); i++ {
					if n.Attr[i].Key == "src" || (n.Data == "link" && n.Attr[i].Key == "href") {
						src := n.Attr[i].Val
						fullSrc := d.resolveURL(src, base)
						if fullSrc != "" {
							d.wg.Add(1)
							go d.download(fullSrc, depth+1)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

// resolveURL resolves relative URLs to absolute.
func (d *Downloader) resolveURL(link string, base string) string {
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}
	baseU, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return baseU.ResolveReference(u).String()
}

// isSameDomain checks if the URL is in the same domain.
func (d *Downloader) isSameDomain(u string) bool {
	parsedU, err := url.Parse(u)
	if err != nil {
		return false
	}
	return parsedU.Host == d.baseURL.Host
}

// urlToLocalPath converts URL to local file path.
func (d *Downloader) urlToLocalPath(u string) (string, error) {
	parsedU, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	path := parsedU.Path
	if path == "" || path == "/" {
		path = "/index.html"
	}

	return filepath.Join(d.dir, parsedU.Host, path), nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: wget [flags] <url>")
		flag.PrintDefaults()
	}

	maxDepth := flag.Int("depth", 3, "maximum recursion depth")
	concurrent := flag.Int("concurrent", 10, "maximum concurrent downloads")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	urlStr := flag.Arg(0)
	dir := "mirror"

	downloader, err := NewDownloader(urlStr, dir, *maxDepth, *concurrent)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := downloader.Download(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Download complete.")
}
