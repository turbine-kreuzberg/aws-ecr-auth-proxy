package lib

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"
)

//go:embed containerd.toml.tmpl
var containerdFS embed.FS

func InstallContainerdConfiguration(ctx context.Context, port int, prefix string) error {
	svc, err := ecrClient(ctx, prefix)
	if err != nil {
		return err
	}

	rules, err := svc.DescribePullThroughCacheRules(ctx)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		err = writeContainerdConfiguration(*rule.UpstreamRegistryUrl, *rule.EcrRepositoryPrefix, port)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeContainerdConfiguration(upstream, prefix string, port int) error {
	// Define the data to pass to the template
	data := struct {
		Port   int
		Prefix string
	}{
		Port:   port,
		Prefix: prefix,
	}

	// fix docker quirks
	if upstream == "registry-1.docker.io" {
		upstream = "docker.io"
	}

	// Create directory if missing
	dir := fmt.Sprintf("/etc/containerd/certs.d/%s", upstream)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create configuration directory: %s", err)
	}

	// Parse the template file
	tmpl, err := template.ParseFS(containerdFS, "containerd.toml.tmpl")
	if err != nil {
		return fmt.Errorf("parsing the template: %s", err)
	}

	path := fmt.Sprintf("%s/hosts.toml", dir)

	// Create the output file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("open configfile for writing: %s", err)
	}
	defer file.Close()

	// Execute the template and write the result to the file
	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("render template to file: %s", err)
	}

	log.Printf("containerd config written to %s\n", path)

	return nil
}
