package exporter

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type RunAs int64

const (
	Script RunAs = iota
	Rest
)

func (r RunAs) String() string {
	switch r {
	case Script:
		return "Script"
	case Rest:
		return "REST"
	}
	return "unknown"
}
func RunAsFromString(runtimeMode string) RunAs {
	switch strings.ToLower(runtimeMode) {
	case "script":
		return Script
	case "rest":
		return Rest
	}
	return Script
}

type RunIn int64

const (
	VM RunIn = iota
	Container
)

func (r RunIn) String() string {
	switch r {
	case VM:
		return "VM"
	case Container:
		return "Container"
	}
	return "unknown"
}
func RunInFromString(runtimeMode string) RunIn {
	switch strings.ToLower(runtimeMode) {
	case "vm":
		return VM
	case "container":
		return Container
	}
	return VM
}

type ContentType int64

const (
	Text ContentType = iota
	CSV
	JSON
	YAML
)

func (t ContentType) String() string {
	switch t {
	case Text:
		return "text"
	case CSV:
		return "CSV"
	case JSON:
		return "JSON"
	case YAML:
		return "YAML"
	}
	return "unknown"
}
func (t ContentType) Suffix() string {
	switch t {
	case Text:
		return "txt"
	case CSV:
		return "csv"
	case JSON:
		return "json"
	case YAML:
		return "yaml"
	}
	return "unknown"
}
func (t ContentType) HttpContentType() string {
	switch t {
	case Text:
		return "application/text"
	case CSV:
		return "text/csv"
	case JSON:
		return "application/json"
	case YAML:
		return "text/yaml"
	}
	return "unknown"
}
func ContentTypeFromString(contentType string) ContentType {
	switch strings.ToLower(contentType) {
	case "text":
		return Text
	case "csv":
		return CSV
	case "json":
		return JSON
	case "yaml":
		return YAML
	}
	return Text
}

type Config struct {
	runAs RunAs
	runIn RunIn

	serverPort        int
	logLevel          string
	namespaceSelector string
	contentType       ContentType
	outputFileName    string
	withResources     bool
}

func NewConfig() *Config {
	config := Config{}

	config.runAs = Script
	config.runIn = VM

	config.logLevel = "info"
	config.namespaceSelector = ""
	config.contentType = Text
	config.outputFileName = "output"

	config.initFromFlags()
	config.initFromEnvVars()

	return &config
}

func (c *Config) initFromFlags() {
	asService := flag.Bool("as-service", false, "Run as REST service")
	flag.IntVar(&c.serverPort, "server-port", 8080, "Server port (only for REST service mode)")
	flag.StringVar(&c.logLevel, "log-level", "info", "Log level, one of debug, info, warn")
	flag.StringVar(&c.namespaceSelector, "ns-selector", "", "Namespace selector, like label1=value1,label2=value2")
	contentType := flag.String("content-type", "text", "Content type, one of text, CSV")
	outputFileName := flag.String("output", "", "Output file name, default is output.<content-type>. File suffix is automatically added")
	flag.BoolVar(&c.withResources, "with-resources", false, "Include resource configuration and usage")
	flag.Parse()

	if *asService {
		c.runAs = Rest
	}
	c.contentType = ContentTypeFromString(*contentType)
	if *outputFileName != "" {
		c.outputFileName = *outputFileName
	}
}

func (c *Config) initFromEnvVars() {
	if _, ok := os.LookupEnv("AS_SERVICE"); ok {
		c.runAs = Rest
	}
	if _, ok := os.LookupEnv("IN_CONTAINER"); ok {
		c.runIn = Container
	}
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		c.logLevel = v
	}
	if v, ok := os.LookupEnv("NS_SELECTOR"); ok {
		c.namespaceSelector = v
	}
	if v, ok := os.LookupEnv("CONTENT_TYPE"); ok {
		c.contentType = ContentTypeFromString(v)
	}
	if v, ok := os.LookupEnv("SERVER_PORT"); ok {
		var err error
		c.serverPort, err = strconv.Atoi(v)
		if err != nil {
			log.Fatalf("Cannot parse SERVER_PORT variable %s", v)
		}
	}
}

func (c *Config) String() string {
	serverPort := strconv.Itoa(c.serverPort)
	if c.RunAsScript() {
		serverPort = "NA"
	}
	return fmt.Sprintf("Run as: %s, Run in: %v, Server port: %s, Log level: %s, Namespace selector: \"%s\", Content type: %s, Output filename: %s, With resources: %v\n",
		c.runAs, c.runIn, serverPort, c.logLevel, c.namespaceSelector, c.contentType, c.outputFileName, c.withResources)
}
func (c *Config) RunAsScript() bool {
	return c.runAs == Script
}
func (c *Config) RunAsService() bool {
	return c.runAs == Rest
}
func (c *Config) RunInVM() bool {
	return c.runIn == VM
}
func (c *Config) RunInContainer() bool {
	return c.runIn == Container
}

func (c *Config) LogLevel() string {
	return c.logLevel
}
func (c *Config) NamespaceSelector() string {
	return c.namespaceSelector
}
func (c *Config) ServerPort() int {
	return c.serverPort
}
func (c *Config) ContentType() ContentType {
	return c.contentType
}
func (c *Config) OutputFileName() string {
	return c.outputFileName
}
func (c *Config) WithResources() bool {
	return c.withResources
}
