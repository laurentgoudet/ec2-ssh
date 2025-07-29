package ec2ssh

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type SSMConfig struct {
	TagKey   string `mapstructure:"tag_key"`
	TagValue string `mapstructure:"tag_value"` // empty means any value
	Command  string `mapstructure:"command"`
}

type Options struct {
	Regions         []string
	UsePrivateIp    bool
	Template        string
	PreviewTemplate string
	Filters         []string
	Profile         string
	PrintOnly       bool
	SSM             SSMConfig `mapstructure:"ssm"`
}

func ParseOptions() Options {
	// Handle completion modes first
	if len(os.Args) > 1 && os.Args[1] == "--completion" {
		printProfileCompletion()
		os.Exit(0)
	}
	
	if len(os.Args) > 1 && os.Args[1] == "--completion-list" {
		profiles := getAWSProfiles()
		for _, profile := range profiles {
			fmt.Println(profile)
		}
		os.Exit(0)
	}
	
	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	// Handle positional profile argument
	var positionalProfile string
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		positionalProfile = os.Args[1]
		// Remove the profile from args so pflag doesn't see it
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$HOME/.config/ec2-ssh")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			panic(err)
		}
	}

	pflag.StringSlice("region", []string{"us-east-1"}, "The AWS region")
	pflag.Bool("use-private-ip", true, "Use private IP instead of public DNS")
	pflag.StringSlice("filters", []string{}, "Filters to apply with the ec2 api call")
	pflag.Bool("print-only", false, "Print connection details only, don't SSH")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.RegisterAlias("UsePrivateIp", "use-private-ip")
	viper.RegisterAlias("regions", "region")

	viper.SetDefault("Region", "us-east-1")
	viper.SetDefault("UsePrivateIp", true)
	viper.SetDefault("Template", `{{ .InstanceId }}: {{index .Tags "Name"}}`)
	viper.SetDefault("PreviewTemplate", `
			Instance Id: {{.InstanceId}}
			Name:        {{index .Tags "Name"}}
			Private IP:  {{.PrivateIpAddress}}
			Public IP:   {{.PublicIpAddress}}

			Tags:
			{{ range $key, $value := .Tags }}
				{{ indent 2 $key }}: {{ $value }}
			{{- end -}}
		`,
	)
	
	// SSM defaults
	viper.SetDefault("ssm.command", "bash -l")

	// Use positional profile if provided
	profile := positionalProfile

	// Auto-detect region from profile if not specified
	regions := viper.GetStringSlice("Regions")
	if len(regions) == 1 && regions[0] == "us-east-1" && profile != "" {
		if detectedRegion := getRegionFromProfile(profile); detectedRegion != "" {
			regions = []string{detectedRegion}
		}
	}

	return Options{
		Regions:         regions,
		UsePrivateIp:    viper.GetBool("UsePrivateIp"),
		Template:        viper.GetString("Template"),
		PreviewTemplate: viper.GetString("PreviewTemplate"),
		Filters:         viper.GetStringSlice("Filters"),
		Profile:         profile,
		PrintOnly:       viper.GetBool("print-only"),
		SSM: SSMConfig{
			TagKey:   viper.GetString("ssm.tag_key"),
			TagValue: viper.GetString("ssm.tag_value"),
			Command:  viper.GetString("ssm.command"),
		},
	}
}

// printProfileCompletion prints a complete bash completion script
func printProfileCompletion() {
	fmt.Print(`#!/bin/bash

# Bash completion for ec2-ssh
_ec2_ssh_completion() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # If we're completing the first argument (profile)
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        local profiles
        profiles=$(ec2-ssh --completion-list 2>/dev/null)
        COMPREPLY=($(compgen -W "$profiles" -- "$cur"))
    fi
}

# Register completion for ec2-ssh
complete -F _ec2_ssh_completion ec2-ssh

# If you want to use 's' as an alias, uncomment this line:
# complete -F _ec2_ssh_completion s
`)
}

// getAWSProfiles extracts profile names from AWS config file
func getAWSProfiles() []string {
	configPath := filepath.Join(os.Getenv("HOME"), ".aws", "config")
	file, err := os.Open(configPath)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var profiles []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[profile ") && strings.HasSuffix(line, "]") {
			profile := strings.TrimPrefix(line, "[profile ")
			profile = strings.TrimSuffix(profile, "]")
			profiles = append(profiles, profile)
		}
	}
	return profiles
}

// getRegionFromProfile extracts region from AWS config for a specific profile
func getRegionFromProfile(profile string) string {
	configPath := filepath.Join(os.Getenv("HOME"), ".aws", "config")
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	var currentProfile string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Check for profile section
		if strings.HasPrefix(line, "[profile ") && strings.HasSuffix(line, "]") {
			currentProfile = strings.TrimPrefix(line, "[profile ")
			currentProfile = strings.TrimSuffix(currentProfile, "]")
			continue
		}
		
		// Check for region in the current profile
		if currentProfile == profile && strings.HasPrefix(line, "region") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
		
		// Reset current profile if we hit a new section
		if strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[profile ") {
			currentProfile = ""
		}
	}
	return ""
}

// formatProfiles formats a list of profiles for display
func formatProfiles(profiles []string) string {
	if len(profiles) == 0 {
		return "none found"
	}
	result := ""
	for i, profile := range profiles {
		if i > 0 {
			result += ", "
		}
		result += profile
		if i >= 4 { // Show first 5 profiles
			remaining := len(profiles) - i - 1
			if remaining > 0 {
				result += fmt.Sprintf(" (and %d more)", remaining)
			}
			break
		}
	}
	return result
}
