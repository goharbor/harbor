package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	vars := map[string]string{}

	targetKeys := map[string]bool{
		"GOBUILDIMAGE":        true,
		"SWAGGER_VERSION":     true,
		"REGISTRY_SRC_TAG":    true,
		"TRIVYVERSION":        true,
		"TRIVYADAPTERVERSION": true,
		"DISTRIBUTION_SRC":    true,
		"SPECTRAL_VERSION":    true,
		"NODEBUILDIMAGE":      true,
	}

	f, err := os.Open("Makefile")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	re := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*[:]?=\s*(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := re.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		key, val := m[1], m[2]
		if !targetKeys[key] {
			continue
		}
		val = strings.Trim(val, `"`)
		vars[key] = strings.TrimSpace(val)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	var goVersion string
	if img, ok := vars["GOBUILDIMAGE"]; ok {
		parts := strings.SplitN(img, ":", 2)
		if len(parts) == 2 {
			goVersion = parts[1]
		}
	}

	var spectralVersion string
	if version, ok := vars["SPECTRAL_VERSION"]; ok {
		parts := strings.SplitN(version, "v", 2)
		if len(parts) == 2 {
			spectralVersion = parts[1]
		}
	}

	var nodeVersion string
	if img, ok := vars["NODEBUILDIMAGE"]; ok {
		parts := strings.SplitN(img, ":", 2)
		if len(parts) == 2 {
			nodeVersion = parts[1]
		}
	}

	swaggerVersion := vars["SWAGGER_VERSION"]
	registryTag := vars["REGISTRY_SRC_TAG"]
	trivyVer := vars["TRIVYVERSION"]
	trivyAdapterVer := vars["TRIVYADAPTERVERSION"]
	distributionSrc := vars["DISTRIBUTION_SRC"]

	const (
		NPM_REGISTRY         = "https://registry.npmjs.org"
		DEV_PLATFORM         = "linux/amd64"
		DEV_VERSION          = "dev"
		DEBUG_PORT           = "4001"
		GOLANGCILINT_VERSION = "latest"
		DELVE_VERSION        = "v1.24.1"
		BUN_VERSION          = "1.2.13"
	)

	outputFile := "./.dagger/consts.go"
	out, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Start writing the Go file
	fmt.Fprintln(out, "// Auto generated from Makefile")
	fmt.Fprintf(out, "package main\n\n")
	fmt.Fprintln(out, "const (")

	fmt.Fprintf(out, "\tGO_VERSION           = \"%s\"\n", goVersion)
	fmt.Fprintf(out, "\tSWAGGER_VERSION      = \"%s\"\n", swaggerVersion)
	fmt.Fprintf(out, "\tNPM_REGISTRY         = \"%s\"\n", NPM_REGISTRY)
	fmt.Fprintf(out, "\tDEV_PLATFORM         = \"%s\"\n", DEV_PLATFORM)
	fmt.Fprintf(out, "\tDEV_VERSION          = \"%s\"\n", DEV_VERSION)
	fmt.Fprintf(out, "\tDEBUG_PORT           = \"%s\"\n", DEBUG_PORT)
	fmt.Fprintf(out, "\tGOLANGCILINT_VERSION = \"%s\"\n", GOLANGCILINT_VERSION)
	fmt.Fprintf(out, "\tDELVE_VERSION        = \"%s\"\n", DELVE_VERSION)
	fmt.Fprintf(out, "\tBUN_VERSION          = \"%s\"\n", BUN_VERSION)
	fmt.Fprintf(out, "\tNODE_VERSION         = \"%s\"\n", nodeVersion)
	fmt.Fprintf(out, "\tDISTRIBUTION_SRC     = \"%s\"\n", distributionSrc)
	fmt.Fprintf(out, "\tSPECTRAL_VERSION     = \"%s\"\n", spectralVersion)
	fmt.Fprintf(out, "\tREGISTRY_SRC_TAG     = \"%s\"\n", registryTag)
	fmt.Fprintln(out, ")")

	tvNoPrefix := strings.TrimPrefix(trivyVer, "v")

	fmt.Fprintln(out, "\nvar (")
	fmt.Fprintf(out, "\tTRIVYVERSION               = \"%s\"\n", trivyVer)
	fmt.Fprintf(out, "\tTRIVYADAPTERVERSION        = \"%s\"\n", trivyAdapterVer)
	fmt.Fprintf(out, "\tTRIVY_VERSION_NO_PREFIX    = \"%s\"\n", tvNoPrefix)
	fmt.Fprintf(out, "\tTRIVY_DOWNLOAD_URL         = \"https://github.com/aquasecurity/trivy/releases/download/%s/trivy_%s_Linux-64bit.tar.gz\"\n", trivyVer, tvNoPrefix)
	fmt.Fprintf(out, "\tTRIVY_ADAPTER_DOWNLOAD_URL = \"https://github.com/goharbor/harbor-scanner-trivy/archive/refs/tags/%s.tar.gz\"\n", trivyAdapterVer)
	fmt.Fprintln(out, ")")

	fmt.Printf("✅ Successfully wrote constants to %s\n", outputFile)
}
