package lib

import (
	"embed"
	"fmt"
	"log"
	"os"
	"text/template"
)

//go:embed systemd.service.tmpl
var systemdFS embed.FS

func InstallSystemdServiceConfiguraiton(port int, prefix string) error {
	exec, err := os.Executable()
	if err != nil {
		return fmt.Errorf("lookup path of binary: %s", err)
	}

	exec = fmt.Sprintf("%s run --prefix \"%s\"", exec, prefix)

	if port != 432 {
		exec = fmt.Sprintf("%s --port %d", exec, port)
	}

	// Define the data to pass to the template
	data := struct {
		Exec string
	}{
		Exec: exec,
	}

	// Parse the template file
	tmpl, err := template.ParseFS(systemdFS, "systemd.service.tmpl")
	if err != nil {
		return fmt.Errorf("parsing the template: %s", err)
	}

	path := "/etc/systemd/system/aws-image-proxy.service"

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
