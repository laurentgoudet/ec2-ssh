package ec2ssh

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Tags []types.Tag

func (s Tags) Len() int {
	return len(s)
}

func (s Tags) Less(i, j int) bool {
	return *s[i].Key < *s[j].Key
}

func (s Tags) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
