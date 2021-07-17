package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type restItem struct {
	// Name is the human-readable name for this item. It should be written lower case and with spaces.
	Name string
	// ID is the SDK identifier for this item. IT must be capitalized.
	ID string
}

func main() {
	name, id, tplDir, targetDir, nofmt, nolint := getParameters()

	name = strings.TrimSpace(name)
	if name == "" {
		_, _ = fmt.Fprintf(os.Stderr, "The -n parameter is required.\n\n")
		flag.Usage()
	}
	id = strings.TrimSpace(id)
	if id == "" {
		_, _ = fmt.Fprintf(os.Stderr, "The -i parameter is required.\n\n")
		flag.Usage()
	}

	restItem := restItem{
		name,
		id,
	}
	if err := filepath.Walk(
		tplDir, func(fn string, info os.FileInfo, _ error) error {
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(fn, ".tpl") {
				return nil
			}
			return handleTemplateFile(fn, id, targetDir, restItem, nofmt)
		},
	); err != nil {
		log.Fatalln(err)
	}
	if !nolint {
		if err := runGoLint(targetDir); err != nil {
			log.Fatalf(
				"Failed to run golangci-lint. You can skip this step by passing -nolint in the command line or setting the NOLINT environment variable. (%v)",
				err,
			)
		}
	}
}

func getParameters() (string, string, string, string, bool, bool) {
	name := ""
	id := ""
	tplDir := "./codetemplates"
	targetDir := "./"
	watch := false
	nofmt := false
	nolint := false
	setupFlags(&name, &id, &tplDir, &targetDir, &watch, &nofmt, &nolint)
	flag.Usage = func() {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Usage: go run rest.go OPTIONS\n\nThis file generates REST client calls based on the templates in the \"codetemplates\" directory.\n",
		)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	if os.Getenv("NOFMT") != "" {
		nofmt = true
	}
	if os.Getenv("NOLINT") != "" {
		nolint = true
	}
	return name, id, tplDir, targetDir, nofmt, nolint
}

func setupFlags(name *string, id *string, tplDir *string, targetDir *string, watch *bool, nofmt *bool, nolint *bool) {
	flag.StringVar(
		name,
		"n",
		"",
		"Pass a human-readable name. E.g. \"storage domain\". Required.",
	)
	flag.StringVar(
		id,
		"i",
		"",
		"Pass an identifier used in the SDK. Must be capitalized. E.g. \"StorageDomain\". Required.",
	)
	flag.StringVar(
		tplDir,
		"d",
		*tplDir,
		fmt.Sprintf(
			"Specify a directory for the source templates. Defaults to \"%s\".",
			*tplDir,
		),
	)
	flag.StringVar(
		targetDir,
		"t",
		*targetDir,
		fmt.Sprintf(
			"Specify a target directory the generated files should be written into. Defaults to \"%s\"",
			*targetDir,
		),
	)
	flag.BoolVar(
		watch,
		"w",
		false,
		"Enable watching templates for changes and update.",
	)
	flag.BoolVar(
		nofmt,
		"nofmt",
		false,
		"Do not run gofmt on resulting file.",
	)
	flag.BoolVar(
		nolint,
		"nolint",
		false,
		"Do not run go golangci-lint on the resulting file.",
	)
}

func handleTemplateFile(templateFileName string, id string, targetDir string, restItem restItem, nofmt bool) error {
	fh, err := os.Open(templateFileName)
	if err != nil {
		return fmt.Errorf("failed to open %s (%w)", templateFileName, err)
	}
	defer func() {
		_ = fh.Close()
	}()
	data, err := ioutil.ReadAll(fh)
	if err != nil {
		return fmt.Errorf("failed to read %s (%w)", templateFileName, err)
	}
	targetFileName := path.Base(filepath.ToSlash(strings.TrimSuffix(templateFileName, ".tpl")))
	file := fmt.Sprintf(strings.ReplaceAll(targetFileName, "ITEM", "%s"), strings.ToLower(id))
	t := path.Join(targetDir, file)
	if err := renderTemplate(string(data), t, restItem); err != nil {
		return err
	}
	if !nofmt {
		if err := runGoFmt(t); err != nil {
			return fmt.Errorf(
				"failed to run go fmt on %s. You can skip this step by passing -nofmt in the command line or setting the NOFMT environment variable. (%w)",
				t,
				err,
			)
		}
	}
	return nil
}

func runGoFmt(t string) error {
	cmd := exec.Command("gofmt", "-w", t)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runGoLint(t string) error {
	cmd := exec.Command("golangci-lint", "run", t)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func renderTemplate(tplText string, file string, restItem restItem) error {
	tpl, err := template.New("list").Funcs(
		map[string]interface{}{
			"toLower": func(input string) string {
				if len(input) == 0 {
					return input
				}
				return fmt.Sprintf("%s%s", strings.ToLower(input[:1]), input[1:])
			},
		},
	).Parse(tplText)
	if err != nil {
		return fmt.Errorf("failed to parse list template (%w)", err)
	}
	fh, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to open %s (%w)", file, err)
	}
	defer func() {
		if err := fh.Close(); err != nil {
			panic(fmt.Errorf("failed to close %s (%w)", file, err))
		}
	}()
	if err := tpl.Execute(fh, restItem); err != nil {
		return fmt.Errorf("failed to render list template to %s (%w)", file, err)
	}
	return nil
}
