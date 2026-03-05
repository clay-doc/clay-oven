package main

import "fmt"

// OutputSink abstracts all output and confirmation operations so the
// pipeline can run identically under the interactive TUI or plain CI mode.
type OutputSink interface {
	Banner()
	Header(msg string)
	Success(msg string)
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	KeyVal(key, val string)
	StructLine(line string)
	Verbose(msg string)
	Confirm(prompt string) bool
	DownloadProgress(received, total int64)
	Done()
}

// ---------------------------------------------------------------------------
// CISink – plain-text output for CI pipelines, auto-confirms everything.
// ---------------------------------------------------------------------------

// CISink writes plain log lines to stdout with no ANSI styling.
type CISink struct{}

func (c *CISink) Banner() {
	fmt.Println("=== Clay Oven ===")
}

func (c *CISink) Header(msg string) {
	fmt.Println()
	fmt.Printf("--- %s ---\n", msg)
}

func (c *CISink) Success(msg string) {
	fmt.Printf("[OK]   %s\n", msg)
}

func (c *CISink) Error(msg string) {
	fmt.Printf("[ERR]  %s\n", msg)
}

func (c *CISink) Warn(msg string) {
	fmt.Printf("[WARN] %s\n", msg)
}

func (c *CISink) Info(msg string) {
	fmt.Printf("[INFO] %s\n", msg)
}

func (c *CISink) KeyVal(key, val string) {
	fmt.Printf("  %-20s -> %s\n", key, val)
}

func (c *CISink) StructLine(line string) {
	fmt.Print(line)
}

func (c *CISink) Verbose(msg string) {
	if Verbose {
		fmt.Printf("[DBG]  %s\n", msg)
	}
}

// Confirm always returns true in CI mode (auto-confirm).
func (c *CISink) Confirm(prompt string) bool {
	fmt.Printf("[AUTO] %s -> yes\n", prompt)
	return true
}

// DownloadProgress prints a percentage line. Only prints at 25% intervals
// and on completion to avoid flooding CI logs.
func (c *CISink) DownloadProgress(received, total int64) {
	if total > 0 {
		pct := float64(received) / float64(total) * 100
		// Only print at 25% milestones and completion to keep CI logs clean.
		milestone := int(pct) / 25
		prevPct := float64(received-1) / float64(total) * 100
		prevMilestone := int(prevPct) / 25
		if milestone != prevMilestone || received >= total {
			fmt.Printf("[DL]   %.0f%% (%d / %d bytes)\n", pct, received, total)
		}
	} else {
		// Unknown total — print sparingly.
		if received%(1024*256) == 0 || received == 0 {
			fmt.Printf("[DL]   %d bytes downloaded\n", received)
		}
	}
}

func (c *CISink) Done() {
	fmt.Println()
	fmt.Println("[DONE] Clay Oven finished.")
}
