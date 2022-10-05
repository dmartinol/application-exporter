package config

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
	Monitoring
)

func (r RunAs) String() string {
	switch r {
	case Script:
		return "Script"
	case Rest:
		return "REST"
	case Monitoring:
		return "Monitoring"
	}
	return "unknown"
}
func RunAsFromString(runtimeMode string) RunAs {
	switch strings.ToLower(runtimeMode) {
	case "script":
		return Script
	case "rest":
		return Rest
	case "monitoring":
		return Monitoring
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

	serverPort    int
	logLevel      string
	burst         int
	contentType   ContentType
	withResources bool

	runnerConfig *RunnerConfig
}

type RunnerConfig struct {
	environment       string
	namespaceSelector string

	outputFileName string
}

func NewConfig() *Config {
	config := Config{}

	config.runAs = Script
	config.runIn = VM

	config.logLevel = "info"
	config.contentType = Text
	config.withResources = false

	config.runnerConfig = NewRunnerConfig()

	config.initFromFlags()
	config.initFromEnvVars()

	return &config
}

func NewRunnerConfig() *RunnerConfig {
	runnerConfig := RunnerConfig{}
	runnerConfig.environment = "default"
	runnerConfig.namespaceSelector = ""
	runnerConfig.outputFileName = "output"
	return &runnerConfig
}

func (c *Config) initFromFlags() {
	asService := flag.Bool("as-service", false, "Run as REST service")
	isMonitoring := flag.Bool("is-monitoring", false, "Run as REST service for Prometheus scraping")
	flag.IntVar(&c.serverPort, "server-port", 8080, "Server port (only for REST service mode)")
	flag.StringVar(&c.logLevel, "log-level", "info", "Log level, one of debug, info, warn")
	contentType := flag.String("content-type", "text", "Content type, one of text, CSV")
	c.contentType = ContentTypeFromString(*contentType)
	flag.BoolVar(&c.withResources, "with-resources", false, "Include resource configuration and usage")
	flag.IntVar(&c.burst, "burst", 40, "Maximum burst for throttle")

	flag.StringVar(&c.runnerConfig.environment, "environment", "default", "Environment name (to tag Prometheus metrics)")
	flag.StringVar(&c.runnerConfig.namespaceSelector, "ns-selector", "", "Namespace selector, like label1=value1,label2=value2")
	outputFileName := flag.String("output", "", "Output file name, default is output.<content-type>. File suffix is automatically added")
	flag.Parse()

	if *asService {
		c.runAs = Rest
	} else if *isMonitoring {
		c.runAs = Monitoring
	}
	if *outputFileName != "" {
		c.runnerConfig.outputFileName = *outputFileName
	}
}

func (c *Config) initFromEnvVars() {
	if _, ok := os.LookupEnv("AS_SERVICE"); ok {
		c.runAs = Rest
	}
	if _, ok := os.LookupEnv("IN_CONTAINER"); ok {
		c.runIn = Container
	}
	if _, ok := os.LookupEnv("IS_MONITORING"); ok {
		c.runAs = Monitoring
	}
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		c.logLevel = v
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

	if v, ok := os.LookupEnv("ENVIRONMENT"); ok {
		c.runnerConfig.environment = v
	}
	if v, ok := os.LookupEnv("NS_SELECTOR"); ok {
		c.runnerConfig.namespaceSelector = v
	}
}

func (c *Config) String() string {
	serverPort := strconv.Itoa(c.serverPort)
	if c.RunAsScript() {
		serverPort = "NA"
	}
	return fmt.Sprintf("Run as: %s, Run in: %v,  Server port: %s, Log level: %s, , Content type: %s, With resources: %v, Burst: %d",
		c.runAs, c.runIn, serverPort, c.logLevel, c.contentType, c.withResources, c.burst)
}
func (c *Config) RunAsScript() bool {
	return c.runAs == Script
}
func (c *Config) RunAsService() bool {
	return c.runAs == Rest
}
func (c *Config) RunAsMonitoring() bool {
	return c.runAs == Monitoring
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
func (c *Config) ServerPort() int {
	return c.serverPort
}
func (c *Config) ContentType() ContentType {
	return c.contentType
}
func (c *Config) WithResources() bool {
	return c.withResources
}
func (c *Config) Burst() int {
	return c.burst
}

func (c *Config) SetContentType(contentType ContentType) {
	c.contentType = contentType
}
func (c *Config) SetWithResources(withResources bool) {
	c.withResources = withResources
}
func (c *Config) SetBurst(burst int) {
	c.burst = burst
}

func (c *Config) GlobalRunnerConfig() *RunnerConfig {
	return c.runnerConfig
}

func (c *RunnerConfig) Environment() string {
	return c.environment
}
func (c *RunnerConfig) NamespaceSelector() string {
	return c.namespaceSelector
}
func (c *RunnerConfig) OutputFileName() string {
	return c.outputFileName
}

func (c *RunnerConfig) SetNamespaceSelector(namespaceSelector string) {
	c.namespaceSelector = namespaceSelector
}
func (c *RunnerConfig) SetOutputFileName(outputFileName string) {
	c.outputFileName = outputFileName
}

func (r *RunnerConfig) String() string {
	return fmt.Sprintf("Environment: %s, Namespace selector: \"%s\", Output filename: %s", r.environment, r.namespaceSelector, r.outputFileName)
}
