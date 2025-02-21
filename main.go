package main

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"crypto/rand"
	"encoding/hex"
	"flag"
)

// Declare a package-level debug flag.
var debug = flag.Bool("debug", false, "enable debug output")

func printFileContents(filename string) error {
/*
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Failed to read file:", filename, ":", err)
	}
	fmt.Println("File contents:", hex.EncodeToString(contents))
*/
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	// Create a hex dumper that writes to stdout
	dumper := hex.Dumper(os.Stdout)
	defer dumper.Close()

	// Copy the file contents to the dumper
	_, err = io.Copy(dumper, file)
	if err != nil {
		fmt.Printf("Error dumping hex: %v\n", err)
		os.Exit(1)
	}
	return nil
}
	
func shred(filename string) error {
	if *debug { printFileContents(filename) }
	// Perform three overwrite passes
	for pass := 0; pass < 3; pass++ {
		// Open file for read-write
		f, err := os.OpenFile(filename, os.O_RDWR, 0)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()

		// Get the size of the file
		fi, err := f.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat file: %w", err)
		}
		size := fi.Size()
		if size == 0 {
			return fmt.Errorf("file is empty")
		}

		bufSize := 4096
		buf := make([]byte, bufSize)

		bytesRemaining := size
		for bytesRemaining > 0 {
			// Determine the number of bytes to write in this iteration
			chunkSize := int64(bufSize)
			if bytesRemaining < chunkSize {
				chunkSize = bytesRemaining
			}

			// Fill buffer with cryptographically secure random data
			if _, err := rand.Read(buf[:chunkSize]); err != nil {
				return fmt.Errorf("failed to generate random data: %w", err)
			}

			// Write buffer to file
			if _, err := f.Write(buf[:chunkSize]); err != nil {
				return fmt.Errorf("failed to write data: %w", err)
			}

			bytesRemaining -= chunkSize
		}

		// Flush writes to disk
		if err := f.Sync(); err != nil {
			return fmt.Errorf("failed to sync data: %w", err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
		if *debug { printFileContents(filename) }
	}

	// Remove the file from disk
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

func testShred() {
	// 1. Create a test file with known content
	testFilename := "testfile.txt"
	initialContent := "Sensitive data that needs to be shredded."
	err := os.WriteFile(testFilename, []byte(initialContent), 0644)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	// 2. Verify that the content before shredding is correct
	contentBefore, err := ioutil.ReadFile(testFilename)
	if string(contentBefore) != initialContent {
		fmt.Println("Error: file content doesn't match initial data before shredding.")
		return
	}

	// 3. Shred the file
	err = shred(testFilename)
	if err != nil {
		fmt.Println("Error shredding file:", err)
		return
	}

	// 4. Check if the file is deleted
	_, err = os.Stat(testFilename)
	if !os.IsNotExist(err) {
		fmt.Println("Error: file still exists after shredding.")
		return
	}

	// 5. (Optional) Attempt to recover content after deletion

	fmt.Println("Test passed. File was shredded and deleted successfully.")
}

func main() {
	flag.Parse()
	// Run the test
	testShred()
}

