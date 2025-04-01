package checks

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
)

// Firewall checks the system firewall.
type Firewall struct {
	passed bool
}

// Name returns the name of the check
func (f *Firewall) Name() string {
	return "Firewall is on"
}

// checkIptables checks if iptables is active
func (f *Firewall) checkIptables() bool {
	output, err := shared.RunCommand("iptables", "-L", "INPUT", "--line-numbers")
	if err != nil {
		log.WithError(err).WithField("output", output).Warn("Failed to check iptables status")
		return false
	}
	log.WithField("output", output).Debug("Iptables status")

	// Define a struct to hold iptables rule information
	type IptablesRule struct {
		Number      int
		Target      string
		Protocol    string
		Options     string
		Source      string
		Destination string
	}

	var rules []IptablesRule
	var policy string

	// Parse the output to check if there are any rules or chains defined
	scanner := bufio.NewScanner(strings.NewReader(output))
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Extract policy from the first line
		if lineCount == 1 && strings.Contains(line, "Chain INPUT") {
			if strings.Contains(line, "policy ACCEPT") {
				policy = "ACCEPT"
			} else if strings.Contains(line, "policy DROP") {
				policy = "DROP"
			} else if strings.Contains(line, "policy REJECT") {
				policy = "REJECT"
			}
			continue
		}

		// Skip the header line
		if lineCount == 2 {
			continue
		}

		// Parse rule lines
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			ruleNum, err := strconv.Atoi(fields[0])
			if err != nil {
				continue // Skip lines that don't start with a number
			}

			rule := IptablesRule{
				Number:      ruleNum,
				Target:      fields[1],
				Protocol:    fields[2],
				Options:     fields[3],
				Source:      fields[4],
				Destination: fields[5],
			}
			rules = append(rules, rule)
		}
	}

	// Check for custom chains like nixos-fw
	hasCustomChain := false
	for _, rule := range rules {
		if rule.Target != "ACCEPT" && rule.Target != "DROP" && rule.Target != "REJECT" {
			hasCustomChain = true
			break
		}
	}

	log.WithField("rules_count", len(rules)).
		WithField("policy", policy).
		WithField("has_custom_chain", hasCustomChain).
		Debug("Iptables has active rules or restrictive policy")

	// Firewall is active if there are rules or the policy is restrictive or custom chains are used
	return len(rules) > 0 || policy == "DROP" || policy == "REJECT" || hasCustomChain
}

// Run executes the check
func (f *Firewall) Run() error {
	f.passed = f.checkIptables()
	return nil
}

// Passed returns the status of the check
func (f *Firewall) Passed() bool {
	return f.passed
}

// IsRunnable returns whether Firewall is runnable.
func (f *Firewall) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (f *Firewall) UUID() string {
	return "2e46c89a-5461-4865-a92e-3b799c12034a"
}

// PassedMessage returns the message to return if the check passed
func (f *Firewall) PassedMessage() string {
	return "Firewall is on"
}

// FailedMessage returns the message to return if the check failed
func (f *Firewall) FailedMessage() string {
	return "Firewall is off"
}

// RequiresRoot returns whether the check requires root access
func (f *Firewall) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *Firewall) Status() string {
	if f.Passed() {
		return f.PassedMessage()
	}

	return f.FailedMessage()
}
