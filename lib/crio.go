package lib

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"
)

//go:embed crio.toml.tmpl
var crioFS embed.FS

type Mirror struct {
	Domain string
	Prefix string
}

func InstallCrioConfiguraiton(ctx context.Context, port int, prefix string) error {
	svc, err := ecrClient(prefix)
	if err != nil {
		return err
	}

	result, err := svc.DescribePullThroughCacheRules(ctx)
	if err != nil {
		return err
	}

	mirrors := []Mirror{}
	for _, rule := range result.PullThroughCacheRules {
		upstream := *rule.UpstreamRegistryUrl

		// fix docker quirks
		if upstream == "registry-1.docker.io" {
			upstream = "docker.io"
		}

		mirrors = append(mirrors, Mirror{
			Domain: upstream,
			Prefix: *rule.EcrRepositoryPrefix,
		})
	}

	err = writeCrioConfiguration(mirrors, port)
	if err != nil {
		return err
	}

	return nil
}

func writeCrioConfiguration(mirrors []Mirror, port int) error {
	// Define the data to pass to the template
	data := struct {
		Port    int
		Mirrors []Mirror
	}{
		Port:    port,
		Mirrors: mirrors,
	}

	// Create directory if missing
	dir := "/etc/containers/registries.conf.d"

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create configuration directory: %s", err)
	}

	// Parse the template file
	tmpl, err := template.ParseFS(crioFS, "crio.toml.tmpl")
	if err != nil {
		return fmt.Errorf("parsing the template: %s", err)
	}

	path := fmt.Sprintf("%s/local-mirrors.conf", dir)

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

	log.Printf("crio config written to %s\n", path)

	return nil
}
