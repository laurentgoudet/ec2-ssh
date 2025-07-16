package ec2ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	finder "github.com/ktr0731/go-fuzzyfinder"
)

type Ec2ssh struct {
	fzfInput        *bytes.Buffer
	options         Options
	listTemplate    *template.Template
	previewTemplate *template.Template
	ec2Clients      []*ec2.Client
}

func New() (*Ec2ssh, error) {
	options := ParseOptions()

	// Check if we have a profile or valid default credentials
	if options.Profile == "" {
		// Try to load default config and test credentials
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("no AWS profile specified and no default credentials found.\n\nUsage:\n  ec2-ssh <profile>  # Use a specific profile\n\nAvailable profiles: %s", 
				formatProfiles(getAWSProfiles()))
		}
		
		// Test if credentials actually work by trying to get caller identity
		_, err = cfg.Credentials.Retrieve(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("no AWS profile specified and default credentials are invalid.\n\nUsage:\n  ec2-ssh <profile>  # Use a specific profile\n\nAvailable profiles: %s", 
				formatProfiles(getAWSProfiles()))
		}
	}

	clients := make([]*ec2.Client, 0)
	for _, region := range options.Regions {
		var cfg aws.Config
		var err error
		
		if options.Profile != "" {
			cfg, err = config.LoadDefaultConfig(context.TODO(), 
				config.WithRegion(region),
				config.WithSharedConfigProfile(options.Profile))
		} else {
			cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		}
		
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		client := ec2.NewFromConfig(cfg)
		clients = append(clients, client)
	}

	tmpl, err := template.New("Instance").Funcs(sprig.TxtFuncMap()).Parse(options.Template)
	if err != nil {
		panic(err)
	}

	previewTemplate, err := template.New("Preview").Funcs(sprig.TxtFuncMap()).Parse(options.PreviewTemplate)
	if err != nil {
		panic(err)
	}

	return &Ec2ssh{
		fzfInput:        new(bytes.Buffer),
		options:         options,
		listTemplate:    tmpl,
		previewTemplate: previewTemplate,
		ec2Clients:      clients,
	}, nil
}

func (e *Ec2ssh) Run() {
	instances := make([]types.Instance, 0)
	instancesLock := &sync.Mutex{}

	wg := &sync.WaitGroup{}
	for _, client := range e.ec2Clients {
		wg.Add(1)
		go func(c *ec2.Client) {
			defer wg.Done()
			retrivedInstances, err := e.ListInstances(c)
			if err != nil {
				panic(err)
			}

			instancesLock.Lock()
			instances = append(instances, retrivedInstances...)
			instancesLock.Unlock()
		}(client)
	}

	wg.Wait()

	indexes, err := finder.FindMulti(
		instances,
		func(i int) string {
			str, _ := TemplateForInstance(&instances[i], e.listTemplate)
			return fmt.Sprintf("%s\n", str)
		},
		finder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}

			str, _ := TemplateForInstance(&instances[i], e.previewTemplate)

			return str
		}),
	)

	if err != nil {
		if errors.Is(err, finder.ErrAbort) {
			os.Exit(1)
		}
		panic(err)
	}

	// Collect all connection details first
	var connectionDetails []string
	for _, idx := range indexes {
		details := e.GetConnectionDetails(&instances[idx])
		if details == "" {
			fmt.Printf("No connection details available for selected instance %s\n", *instances[idx].InstanceId)
			fmt.Printf("Debug - Public DNS: %v, Public IP: %v, Private IP: %v\n", 
				getStringPtr(instances[idx].PublicDnsName),
				getStringPtr(instances[idx].PublicIpAddress),
				getStringPtr(instances[idx].PrivateIpAddress))
			continue
		}
		connectionDetails = append(connectionDetails, details)
	}

	if len(connectionDetails) == 0 {
		fmt.Println("No valid connection details found")
		os.Exit(1)
	}

	// If print-only flag is set, just print and exit
	if e.options.PrintOnly {
		for _, details := range connectionDetails {
			fmt.Printf("%s\n", details)
		}
		return
	}

	// Automatically use xpanes for multiple instances
	if len(connectionDetails) > 1 {
		fmt.Printf("Connecting to %d instances using xpanes...\n", len(connectionDetails))
		
		// Check if xpanes is available
		if _, err := exec.LookPath("xpanes"); err != nil {
			fmt.Println("Error: xpanes not found. Install with: brew install xpanes")
			fmt.Println("Falling back to single instance connection...")
			
			// Fall back to single instance
			details := connectionDetails[0]
			fmt.Printf("Connecting to %s...\n", details)
			
			cmd := exec.Command("ssh", details)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			err := cmd.Run()
			if err != nil {
				fmt.Printf("SSH connection failed: %v\n", err)
				os.Exit(1)
			}
			return
		}
		
		// Use xpanes to connect to all instances
		args := []string{"-c", "ssh {}"}
		args = append(args, connectionDetails...)
		
		cmd := exec.Command("xpanes", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		err := cmd.Run()
		if err != nil {
			fmt.Printf("xpanes command failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Single instance mode
		details := connectionDetails[0]
		fmt.Printf("Connecting to %s...\n", details)
		
		// Execute SSH command
		cmd := exec.Command("ssh", details)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		err := cmd.Run()
		if err != nil {
			fmt.Printf("SSH connection failed: %v\n", err)
			os.Exit(1)
		}
	}
}

// getStringPtr safely gets string value from pointer
func getStringPtr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
