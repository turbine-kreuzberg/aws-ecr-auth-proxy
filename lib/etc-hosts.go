package lib

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func EtcHostsBlock(ctx context.Context) error {
	svc, err := ecrClient()
	if err != nil {
		return err
	}

	result, err := svc.DescribePullThroughCacheRules(ctx, &ecr.DescribePullThroughCacheRulesInput{})
	if err != nil {
		return err
	}

	hosts := []string{}
	for _, rule := range result.PullThroughCacheRules {
		hosts = append(hosts, *rule.UpstreamRegistryUrl)
	}

	err = addBlockToEtcHosts(hosts)
	if err != nil {
		return err
	}

	return nil
}

func addBlockToEtcHosts(hosts []string) error {
	// Read the /etc/hosts file
	hostsFile := "/etc/hosts"
	file, err := os.Open(hostsFile)
	if err != nil {
		return fmt.Errorf("failed to opening /etc/hosts: %s", err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	existingHosts := make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "0.0.0.0") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				existingHosts[parts[1]] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read /etc/hosts: %s", err)
	}

	// Append new hosts to the file if they don't already exist
	var newLines []string
	for _, host := range hosts {
		if !existingHosts[host] {
			newLines = append(newLines, fmt.Sprintf("0.0.0.0 %s", host))
		}
	}

	if len(newLines) > 0 {
		f, err := os.OpenFile(hostsFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open /etc/hosts for writing: %s", err)
		}
		defer f.Close()

		for _, line := range newLines {
			if _, err = f.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("failed to write to /etc/hosts: %s", err)
			}
		}
	}

	log.Println("host entries written to /etc/hosts")

	return nil
}
