package ec2ssh

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	finder "github.com/ktr0731/go-fuzzyfinder"
)

type Ec2ssh struct {
	fzfInput        *bytes.Buffer
	options         Options
	listTemplate    *template.Template
	previewTemplate *template.Template
	ec2Clients      []*ec2.Client
	ssmClients      []*ssm.Client
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
	ssmClients := make([]*ssm.Client, 0)
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
		
		ssmClient := ssm.NewFromConfig(cfg)
		ssmClients = append(ssmClients, ssmClient)
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
		ssmClients:      ssmClients,
	}, nil
}

func (e *Ec2ssh) Run() {
	instances := make([]types.Instance, 0)
	instancesLock := &sync.Mutex{}
	var lastError error

	wg := &sync.WaitGroup{}
	for _, client := range e.ec2Clients {
		wg.Add(1)
		go func(c *ec2.Client) {
			defer wg.Done()
			retrivedInstances, err := e.ListInstances(c)
			if err != nil {
				instancesLock.Lock()
				lastError = err
				instancesLock.Unlock()
				return
			}

			instancesLock.Lock()
			instances = append(instances, retrivedInstances...)
			instancesLock.Unlock()
		}(client)
	}

	wg.Wait()

	// Handle SSO authentication errors
	if lastError != nil {
		if e.handleSSOError(lastError) {
			// Retry after SSO login
			e.Run()
			return
		}
		panic(lastError)
	}

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
	var ssmConnections []bool
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
		ssmConnections = append(ssmConnections, strings.HasPrefix(details, "ssm:"))
	}

	if len(connectionDetails) == 0 {
		fmt.Println("No valid connection details found")
		os.Exit(1)
	}

	// If print-only flag is set, just print and exit
	if e.options.PrintOnly {
		for i, details := range connectionDetails {
			if ssmConnections[i] {
				instanceId := strings.TrimPrefix(details, "ssm:")
				if e.options.Profile != "" {
					fmt.Printf("aws ssm start-session --target %s --profile %s\n", instanceId, e.options.Profile)
				} else {
					fmt.Printf("aws ssm start-session --target %s\n", instanceId)
				}
			} else {
				fmt.Printf("ssh %s\n", details)
			}
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
			isSSM := ssmConnections[0]
			e.connectToInstance(details, isSSM)
			return
		}
		
		// Use xpanes to connect to all instances
		var args []string
		for i, details := range connectionDetails {
			if ssmConnections[i] {
				instanceId := strings.TrimPrefix(details, "ssm:")
				var command string
				if e.options.Profile != "" {
					command = fmt.Sprintf("aws ssm start-session --target %s --profile %s --document-name AWS-StartInteractiveCommand --parameters 'command=[\"%s\"]'", instanceId, e.options.Profile, e.options.SSM.Command)
				} else {
					command = fmt.Sprintf("aws ssm start-session --target %s --document-name AWS-StartInteractiveCommand --parameters 'command=[\"%s\"]'", instanceId, e.options.SSM.Command)
				}
				args = append(args, command)
			} else {
				args = append(args, fmt.Sprintf("ssh %s", details))
			}
		}
		
		xpanesArgs := []string{"-c", "{}"}
		xpanesArgs = append(xpanesArgs, args...)
		
		cmd := exec.Command("xpanes", xpanesArgs...)
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
		isSSM := ssmConnections[0]
		e.connectToInstance(details, isSSM)
	}
}

func (e *Ec2ssh) connectToInstance(details string, isSSM bool) {
	if isSSM {
		instanceId := strings.TrimPrefix(details, "ssm:")
		fmt.Printf("Connecting to %s via SSM...\n", instanceId)
		
		// Build AWS CLI command with profile if specified
		args := []string{"ssm", "start-session", "--target", instanceId}
		if e.options.Profile != "" {
			args = append(args, "--profile", e.options.Profile)
		}
		args = append(args, "--document-name", "AWS-StartInteractiveCommand")
		args = append(args, "--parameters", fmt.Sprintf("command=[\"%s\"]", e.options.SSM.Command))
		
		cmd := exec.Command("aws", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		err := cmd.Run()
		if err != nil {
			fmt.Printf("SSM connection failed: %v\n", err)
			os.Exit(1)
		}
	} else {
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

// handleSSOError detects SSO authentication errors and automatically runs aws sso login
func (e *Ec2ssh) handleSSOError(err error) bool {
	errStr := err.Error()
	
	// Check if this is an SSO authentication error
	if strings.Contains(errStr, "failed to refresh cached credentials") ||
		strings.Contains(errStr, "cached SSO token") ||
		strings.Contains(errStr, "sso/cache") {
		
		fmt.Printf("SSO session expired. Running 'aws sso login' for profile '%s'...\n", e.options.Profile)
		
		// Get SSO session name from the profile
		ssoSession := e.getSSOSessionFromProfile(e.options.Profile)
		if ssoSession == "" {
			fmt.Printf("Could not determine SSO session for profile '%s'. Please run 'aws sso login --profile %s' manually.\n", e.options.Profile, e.options.Profile)
			return false
		}
		
		// Run aws sso login with the SSO session
		cmd := exec.Command("aws", "sso", "login", "--sso-session", ssoSession)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		err := cmd.Run()
		if err != nil {
			fmt.Printf("SSO login failed: %v\n", err)
			return false
		}
		
		fmt.Println("SSO login successful. Retrying...")
		return true
	}
	
	return false
}

// getSSOSessionFromProfile extracts SSO session name from AWS config for a specific profile
func (e *Ec2ssh) getSSOSessionFromProfile(profile string) string {
	if profile == "" {
		return ""
	}
	
	configPath := filepath.Join(os.Getenv("HOME"), ".aws", "config")
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer file.Close()
	
	var currentProfile string
	var inTargetProfile bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Check for profile section
		if strings.HasPrefix(line, "[profile ") && strings.HasSuffix(line, "]") {
			currentProfile = strings.TrimPrefix(line, "[profile ")
			currentProfile = strings.TrimSuffix(currentProfile, "]")
			inTargetProfile = (currentProfile == profile)
			continue
		}
		
		// Reset if we hit a new section that's not a profile
		if strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[profile ") {
			inTargetProfile = false
			continue
		}
		
		// Look for sso_session in the target profile
		if inTargetProfile && strings.HasPrefix(line, "sso_session") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	
	return ""
}

// getStringPtr safely gets string value from pointer
func getStringPtr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
