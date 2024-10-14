package lib

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestAddBlockToEtcHosts(t *testing.T) {
	// Create a temporary file to simulate /etc/hosts
	tmpfile, err := os.CreateTemp("", "hosts")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some initial content to the temp file
	initialContent := `127.0.0.1 localhost
::1 localhost
0.0.0.0 existing.host.com
`
	if _, err := tmpfile.Write([]byte(initialContent)); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tmpfile.Close()

	// Override the hostsFile variable for testing
	hostsFile = tmpfile.Name()

	// Test adding new hosts
	newHosts := []string{"new1.host.com", "new2.host.com", "existing.host.com"}
	err = addBlockToEtcHosts(newHosts)
	if err != nil {
		t.Fatalf("addBlockToEtcHosts failed: %v", err)
	}

	// Read the file content after adding new hosts
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		t.Fatalf("Failed to read temporary file: %v", err)
	}

	// Check if new hosts were added correctly
	expectedLines := []string{
		"0.0.0.0 new1.host.com",
		"0.0.0.0 new2.host.com",
	}
	for _, line := range expectedLines {
		if !strings.Contains(string(content), line) {
			t.Errorf("Expected line not found: %s", line)
		}
	}

	// Check if existing host was not duplicated
	count := strings.Count(string(content), "0.0.0.0 existing.host.com")
	if count != 1 {
		t.Errorf("Existing host was duplicated, found %d occurrences", count)
	}

	// Check total number of lines
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	expectedLineCount := 5 // 3 initial lines + 2 new lines
	if lineCount != expectedLineCount {
		t.Errorf("Expected %d lines, but found %d", expectedLineCount, lineCount)
	}
}
