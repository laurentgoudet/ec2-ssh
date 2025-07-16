# 🚀 ec2-ssh

ec2-ssh is a command-line tool that provides an interactive fuzzy finder interface for AWS EC2 instances. It utilizes the [fzf](https://github.com/junegunn/fzf) fuzzy matcher to help you quickly search, filter, and select EC2 instances across multiple AWS regions.

![GIF](https://raw.githubusercontent.com/laurentgoudet/ec2-ssh/master/img/ec2-ssh.gif)

## ✨ What's New in v2.0

- **🔐 AWS SSO/Identity Center Support**: Full support for modern AWS authentication
- **⚡ AWS SDK v2**: Updated to the latest AWS SDK for better performance and reliability
- **🎯 Positional Profile Support**: Simply use `ec2-ssh prod` instead of flags
- **🚀 Go 1.22**: Updated to the latest Go version with improved performance
- **🔧 Integrated Completion**: Built-in bash completion script generation
- **🔗 Direct SSH Integration**: Automatically SSHs into selected instances
- **🏠 Private IP Default**: Uses private IP by default for VPC connections
- **🔀 Smart Multi-Instance Support**: Automatically uses xpanes when multiple instances selected

## 📦 Installation

### 🛠️ From Source

```bash
git clone https://github.com/laurentgoudet/ec2-ssh
cd ec2-ssh
go mod download
go build -o ec2-ssh ./cmd/ec2-ssh
```

### 📥 Using Go Install

```bash
go install github.com/laurentgoudet/ec2-ssh/cmd/ec2-ssh@latest
```

## 🎯 Usage

### 🔧 Basic Usage

```bash
# Select an instance and SSH into it (uses private IP by default)
ec2-ssh

# Use with positional profile argument (automatically detects region)
ec2-ssh prod

# Use public DNS/IP instead of private IP
ec2-ssh prod --use-private-ip=false

# Specify a region
ec2-ssh --region us-west-2

# Just print connection details without SSHing (for scripts)
ec2-ssh prod --print-only

# Use in scripts
HOST=$(ec2-ssh prod --print-only)
ssh $HOST

# Use public IP for scripting
HOST=$(ec2-ssh prod --use-private-ip=false --print-only)
ssh $HOST

# Connect to multiple instances - automatically uses xpanes when multiple selected
# (select multiple instances with Tab/Space in the fuzzy finder)
ec2-ssh prod

# Multi-region support
ec2-ssh prod --region us-east-1 --region us-west-2
```

### ⚡ Bash Completion

Set up bash completion for easy profile selection:

```bash
# Source the completion script directly
source <(ec2-ssh --completion)

# Or add to your .bashrc for permanent setup
echo 'source <(ec2-ssh --completion)' >> ~/.bashrc

# Create the alias (uncomment the last line in the completion script for 's' completion)
alias s='ec2-ssh'
```

The completion will suggest available AWS profiles when you type:
```bash
ec2-ssh <TAB>
# or with alias:
s <TAB>
```

**Note:** The `--completion` flag generates a complete bash script that handles all completion logic internally.

### 🔀 Multi-Instance Support

Connect to multiple instances simultaneously - automatically detected:

```bash
# Select multiple instances with Tab/Space in the fuzzy finder
# Automatically uses xpanes when multiple instances are selected
ec2-ssh prod

# Multi-region support - query multiple regions and select instances
ec2-ssh prod --region us-east-1 --region us-west-2

# Print multiple instance IPs for scripting
ec2-ssh prod --print-only
# (then select multiple instances)
```

**Features:**
- **Automatic detection** - no flags needed
- **Graceful fallback** - if xpanes not installed, connects to first instance
- **Smart behavior** - single selection = SSH, multiple = xpanes

**Requirements:**
- Install xpanes for multi-instance support: `brew install xpanes`
- Uses tmux for session management

### 🔍 Filtering

You can filter instances using the `--filters` flag. Use it multiple times to combine filters:

```bash
# Filter by tags
ec2-ssh --filters tag:Environment=production --filters tag:Name=web-server

# Filter by instance state
ec2-ssh --filters instance-state-name=running

# Filter by instance type
ec2-ssh --filters instance-type=t3.micro
```

Valid filter values are those used in the [AWS SDK for Go](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeInstancesInput).

## ⚙️ Configuration

You can set default configuration options in `~/.config/ec2-ssh/config.toml`:

```toml
# Default region
Region = "us-east-1"

# Custom display template
Template = "{{index .Tags \"Name\"}}"

# Use private IP by default (default: true)
UsePrivateIp = true
```

### 🎨 Template Customization

The template uses Go's text/template syntax. Available fields include:
- `.InstanceId` - EC2 instance ID
- `.PublicDnsName` - Public DNS name
- `.PrivateIpAddress` - Private IP address
- `.State.Name` - Instance state
- `.Tags` - Instance tags (use `{{index .Tags "TagName"}}`)

## 📋 Requirements

- 🔧 AWS CLI configured with appropriate credentials (supports AWS SSO/Identity Center)
- 🚀 Go 1.22 or later
- 🔍 [fzf](https://github.com/junegunn/fzf) installed

## 🔐 Authentication

ec2-ssh supports modern AWS authentication methods:

- **🔐 AWS SSO/Identity Center**: Use positional profile argument `ec2-ssh prod`
- **🎭 IAM roles**: For EC2 instances with attached roles
- **🔑 Traditional credentials**: From `~/.aws/credentials` or environment variables
- **🔄 AssumeRole**: Via AWS profiles configured in `~/.aws/config`

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
