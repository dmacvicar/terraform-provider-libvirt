package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// URLStream represents a stream from a URL with its size
type URLStream struct {
	Reader io.ReadCloser
	Size   int64
}

// OpenURLStream opens a URL for reading and returns the stream with its size.
// Supports:
// - https:// URLs (requires Content-Length header)
// - file:// URLs
// - Plain absolute paths (treated as local files)
func OpenURLStream(ctx context.Context, url string) (*URLStream, error) {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return openHTTPStream(ctx, url)
	}

	// Handle file:// URLs and plain paths
	path := url
	if strings.HasPrefix(url, "file://") {
		path = strings.TrimPrefix(url, "file://")
	}

	return openFileStream(ctx, path)
}

// openHTTPStream opens an HTTP(S) URL and returns a stream
func openHTTPStream(ctx context.Context, url string) (*URLStream, error) {
	tflog.Debug(ctx, "Opening HTTP stream", map[string]any{"url": url})

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Require Content-Length header
	if resp.ContentLength < 0 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("server did not provide Content-Length header, cannot determine volume size")
	}

	tflog.Debug(ctx, "HTTP stream opened", map[string]any{
		"url":  url,
		"size": resp.ContentLength,
	})

	return &URLStream{
		Reader: resp.Body,
		Size:   resp.ContentLength,
	}, nil
}

// openFileStream opens a local file and returns a stream
func openFileStream(ctx context.Context, path string) (*URLStream, error) {
	tflog.Debug(ctx, "Opening file stream", map[string]any{"path": path})

	// Get file size
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	tflog.Debug(ctx, "File stream opened", map[string]any{
		"path": path,
		"size": stat.Size(),
	})

	return &URLStream{
		Reader: file,
		Size:   stat.Size(),
	}, nil
}
