package ec2ssh

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func (e *Ec2ssh) ListInstances(ec2Client *ec2.Client) ([]types.Instance, error) {
	instances := make([]types.Instance, 0)
	filters := make([]types.Filter, 0, 0)

	filters = append(filters, types.Filter{
		Name:   aws.String("instance-state-name"),
		Values: []string{"pending", "running", "shutting-down"},
	})

	for _, filter := range e.options.Filters {
		split := strings.SplitN(filter, "=", 2)
		if len(split) < 2 {
			return nil, fmt.Errorf("Filters can only contain one '='. Filter \"%s\" has %d", filter, len(split))
		}

		filters = append(filters, types.Filter{
			Name:   aws.String(split[0]),
			Values: []string{split[1]},
		})
	}
	params := &ec2.DescribeInstancesInput{}

	if len(filters) > 0 {
		params.Filters = filters
	}

	paginator := ec2.NewDescribeInstancesPaginator(ec2Client, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, r := range page.Reservations {
			for _, i := range r.Instances {
				instances = append(instances, i)
			}
		}
	}

	return instances, nil
}

func (e *Ec2ssh) GetConnectionDetails(instance *types.Instance) string {
	// Check if this instance should use SSM
	if e.shouldUseSSM(instance) {
		return "ssm:" + *instance.InstanceId
	}
	
	if e.options.UsePrivateIp {
		if instance.PrivateIpAddress != nil && *instance.PrivateIpAddress != "" {
			return *instance.PrivateIpAddress
		}
		return ""
	}
	
	// Try public DNS first
	if instance.PublicDnsName != nil && *instance.PublicDnsName != "" {
		return *instance.PublicDnsName
	}
	
	// Fall back to public IP
	if instance.PublicIpAddress != nil && *instance.PublicIpAddress != "" {
		return *instance.PublicIpAddress
	}
	
	// Don't fall back to private IP when explicitly not requested
	return ""
}

func (e *Ec2ssh) shouldUseSSM(instance *types.Instance) bool {
	if e.options.SSM.TagKey == "" {
		return false
	}
	
	for _, tag := range instance.Tags {
		if tag.Key != nil && *tag.Key == e.options.SSM.TagKey {
			// If no specific value is required, any value matches
			if e.options.SSM.TagValue == "" {
				return true
			}
			// Otherwise, check for exact match
			if tag.Value != nil && *tag.Value == e.options.SSM.TagValue {
				return true
			}
		}
	}
	return false
}

func TemplateForInstance(i *types.Instance, t *template.Template) (output string, err error) {
	tags := make(map[string]string)

	for _, t := range i.Tags {
		tags[*t.Key] = *t.Value
	}

	buffer := new(bytes.Buffer)
	err = t.Execute(
		buffer,
		struct {
			Tags map[string]string
			*types.Instance
		}{
			tags,
			i,
		},
	)

	output = buffer.String()
	return
}

func InstanceIdFromString(s string) (string, error) {
	i := strings.Index(s, ":")

	if i < 0 {
		return "", fmt.Errorf("Unable to find instance id")
	}
	return strings.TrimSpace(s[0:i]), nil
}
