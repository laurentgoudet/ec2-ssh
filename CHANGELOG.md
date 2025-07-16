# Changelog

## v2.0.0 (2025-01-16)

### 🚀 Major Features

- **Project Renamed**: Renamed from `ec2-fzf` to `ec2-ssh` to better reflect its primary use case
- **AWS SSO/Identity Center Support**: Full support for modern AWS authentication including SSO and Identity Center
- **AWS SDK v2 Migration**: Updated from AWS SDK v1 to v2 for better performance and modern authentication support
- **Positional Profile Support**: Use `ec2-ssh prod` instead of `--profile` flag for cleaner syntax
- **Direct SSH Integration**: Automatically SSHs into selected instances instead of just printing details
- **Smart Multi-Instance Support**: Automatically uses xpanes when multiple instances are selected
- **Integrated Bash Completion**: Built-in completion script generation with `--completion` flag
- **Auto Region Detection**: Automatically detects region from AWS profile configuration

### 🔄 Breaking Changes

- **Minimum Go Version**: Now requires Go 1.22 or later
- **AWS SDK**: Migrated from AWS SDK v1 to v2 (internal change, no API changes)
- **Default Behavior**: Now SSHs directly instead of printing connection details (use `--print-only` for old behavior)
- **Private IP Default**: Uses private IP by default instead of public DNS (use `--use-private-ip=false` for public)

### 🛠️ Improvements

- **Better Error Handling**: Improved error messages for AWS configuration issues and credential validation
- **Modern Dependencies**: Updated to latest versions of all dependencies
- **Performance**: Better concurrent handling of multi-region queries
- **Smart Connection Logic**: Automatic fallback from public DNS → public IP → private IP
- **Multi-Instance UX**: No flags needed - automatically detects multiple selections and uses xpanes
- **Documentation**: Updated README with comprehensive authentication examples and usage patterns

### 🔧 Technical Changes

- Updated `go.mod` to require Go 1.22
- Replaced `github.com/aws/aws-sdk-go` with `github.com/aws/aws-sdk-go-v2`
- Enhanced authentication logic to support multiple credential providers
- Added proper nil checking for optional EC2 instance fields
- Integrated xpanes support for multi-instance connections
- Added AWS config file parsing for profile region detection
- Improved connection detail resolution with smart fallback logic

### 📝 Documentation

- Updated README with comprehensive usage examples
- Added multi-instance support documentation
- Enhanced authentication section with SSO examples
- Updated installation instructions and requirements
- Added bash completion setup guide
- Updated configuration examples for new features

## From 1.0

