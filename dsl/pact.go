/*
Package dsl contains the main Pact DSL used in the Consumer
collaboration test cases, and Provider contract test verification.
*/
package dsl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/dsl/native"
	"github.com/pact-foundation/pact-go/types"
)

// Pact is the container structure to run the Consumer Pact test cases.
type Pact struct {
	// Current server for the consumer.
	Server *types.MockServer `json:"-"`

	// Port the Pact Daemon is running on.
	Port int `json:"-"`

	// Pact RPC Client.
	// pactClient *PactClient

	// Consumer is the name of the Consumer/Client.
	Consumer string `json:"consumer"`

	// Provider is the name of the Providing service.
	Provider string `json:"provider"`

	// Interactions contains all of the Mock Service Interactions to be setup.
	Interactions []*Interaction `json:"interactions"`

	// Log levels.
	LogLevel string `json:"-"`

	// Used to detect if logging has been configured.
	logFilter *logutils.LevelFilter

	// Location of Pact external service invocation output logging.
	// Defaults to `<cwd>/logs`.
	LogDir string `json:"-"`

	// Pact files will be saved in this folder.
	// Defaults to `<cwd>/pacts`.
	PactDir string `json:"-"`

	// Specify which version of the Pact Specification should be used (1 or 2).
	// Defaults to 2.
	SpecificationVersion string `json:"pactSpecificationVersion"`
}

// AddInteraction creates a new Pact interaction, initialising all
// required things. Will automatically start a Mock Service if none running.
func (p *Pact) AddInteraction() *Interaction {
	p.Setup()
	log.Printf("[DEBUG] pact add interaction")
	i := &Interaction{}
	p.Interactions = append(p.Interactions, i)
	return i
}

// Setup starts the Pact Mock Server. This is usually called before each test
// suite begins. AddInteraction() will automatically call this if no Mock Server
// has been started.
func (p *Pact) Setup() *Pact {
	p.setupLogging()
	log.Printf("[DEBUG] pact setup")
	dir, _ := os.Getwd()

	if p.LogDir == "" {
		p.LogDir = fmt.Sprintf(filepath.Join(dir, "logs"))
	}

	if p.PactDir == "" {
		p.PactDir = fmt.Sprintf(filepath.Join(dir, "pacts"))
	}

	if p.SpecificationVersion == "" {
		p.SpecificationVersion = "2.0.0"
	}

	// if p.Server == nil {
	// 	args := []string{
	// 		fmt.Sprintf("--pact-specification-version %d", p.SpecificationVersion),
	// 		fmt.Sprintf("--pact-dir %s", p.PactDir),
	// 		fmt.Sprintf("--log %s/pact.log", p.LogDir),
	// 		fmt.Sprintf("--consumer %s", p.Consumer),
	// 		fmt.Sprintf("--provider %s", p.Provider),
	// 	}
	// 	client := &PactClient{Port: p.Port}
	// 	p.pactClient = client
	// 	// p.Server = client.StartServer()
	// }

	return p
}

// Configure logging
func (p *Pact) setupLogging() {
	if p.logFilter == nil {
		if p.LogLevel == "" {
			p.LogLevel = "INFO"
		}
		p.logFilter = &logutils.LevelFilter{
			Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
			MinLevel: logutils.LogLevel(p.LogLevel),
			Writer:   os.Stderr,
		}
		log.SetOutput(p.logFilter)
	}
	log.Printf("[DEBUG] pact setup logging")
}

// Teardown stops the Pact Mock Server. This usually is called on completion
// of each test suite.
func (p *Pact) Teardown() *Pact {
	log.Printf("[DEBUG] teardown")
	// if p.Server != nil {
	// 	p.Server = p.pactClient.StopServer(p.Server)
	// }
	return p
}

// Verify runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite.
func (p *Pact) Verify(integrationTest func() error) error {
	// p.Setup()
	fmt.Println("CReating mock server from pact file: ", p.formatJSONObject(p))
	port := native.CreateMockServer(p.formatJSONObject(p))
	log.Printf("[DEBUG] pact verify")
	// mockServer := &MockService{
	// 	BaseURL:  fmt.Sprintf("http://localhost:%d", p.Server.Port),
	// 	Consumer: p.Consumer,
	// 	Provider: p.Provider,
	// }

	// Run the integration test
	integrationTest()

	res, mismatches := native.Verify(port, p.PactDir)
	fmt.Println("Result from verify:", res, mismatches)
	if !res {
		return fmt.Errorf("Pact validation failed!")
	}
	// Run Verification Process
	// err := mockServer.Verify()
	// if err != nil {
	// 	return err
	// }

	// Clear out interations
	// p.Interactions = make([]*Interaction, 0)

	// return mockServer.DeleteInteractions()
	return nil
}

// WritePact should be called writes when all tests have been performed for a
// given Consumer <-> Provider pair. It will write out the Pact to the
// configured file.
func (p *Pact) WritePact() error {
	p.Setup()
	log.Printf("[DEBUG] pact write Pact file")
	mockServer := MockService{
		BaseURL:  fmt.Sprintf("http://localhost:%d", p.Server.Port),
		Consumer: p.Consumer,
		Provider: p.Provider,
	}
	err := mockServer.WritePact()
	if err != nil {
		return err
	}

	return nil
}

// VerifyProvider reads the provided pact files and runs verification against
// a running Provider API.
func (p *Pact) VerifyProvider(request types.VerifyRequest) error {
	p.Setup()

	// If we provide a Broker, we go to it to find consumers
	if request.BrokerURL != "" {
		log.Printf("[DEBUG] pact provider verification - finding all consumers from broker: %s", request.BrokerURL)
		err := findConsumers(p.Provider, &request)
		if err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] pact provider verification")

	// content, err := p.pactClient.VerifyProvider(request)
	//
	// // Output test result to screen
	// log.Println(content)
	//
	// return err
	return nil
}

// Format a JSON document to make comparison easier.
func (p *Pact) formatJSON(object string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(object), "", "\t")
	return string(out.Bytes())
}

// Format a JSON document for creating Pact files.
func (p *Pact) formatJSONObject(object interface{}) string {
	out, _ := json.Marshal(object)
	return p.formatJSON(string(out))
}
