# TODO - ec2-ssh Improvements

## 🚀 High Priority

### Performance & Reliability
- [ ] **Caching**: Cache AWS API responses for configurable duration (default: 5 minutes)
  - Cache instance lists per region/profile combination
  - Implement cache invalidation on manual refresh
  - Add `--no-cache` flag for bypassing cache
- [ ] **Custom SSH args**: Support passing SSH flags and options
  - Add `--ssh-opts` flag for custom SSH arguments
  - Support `-i keyfile`, `-p port`, `-o StrictHostKeyChecking=no`, etc.
  - Read SSH config files for host-specific settings
- [ ] **Connection history**: Track and display connection metadata
  - Store last connection time per instance
  - Show connection history in preview pane
  - Add "recently used" sorting option
- [ ] **Homebrew formula**: Create official Homebrew tap for easy installation
  - Set up automated formula updates
  - Add installation instructions to README

### User Experience
- [ ] **Connection pooling**: Reuse AWS SDK clients across regions
- [ ] **Timeout handling**: Add configurable timeouts for AWS API calls and SSH
- [ ] **Retry logic**: Implement exponential backoff for transient failures
- [ ] **Recent instances**: Remember recently connected instances for quick access

## 🔧 Medium Priority

### Configuration & Customization
- [ ] **Instance sorting**: Support multiple sorting criteria
  - Sort by launch time, name, instance type, state
  - Add `--sort-by` flag and config option
- [ ] **Color themes**: Customizable colors for different instance states
  - Different colors for running, stopped, pending instances
  - Support for light/dark themes
- [ ] **Profile aliases**: Short aliases for frequently used profiles
  - Configure in `~/.config/ec2-ssh/aliases.toml`
  - Support tab completion for aliases
- [ ] **Custom actions**: Define custom commands to run on selected instances
  - Configure custom scripts in config file
  - Support for instance-specific variables

### Search & Discovery
- [ ] **Search improvements**: Enhanced fuzzy search capabilities
  - Search within instance names, tags, and IDs
  - Support for regex and glob patterns
  - Highlight matching text in results
- [ ] **Bookmarks**: Save favorite instances with custom names
  - Persistent bookmarks across sessions
  - Quick access to bookmarked instances
- [ ] **Instance filtering**: Advanced filtering options
  - Filter by instance age, size, vpc, availability zone
  - Support for complex filter expressions

## 📊 Low Priority

### Monitoring & Logging
- [ ] **Metrics**: Track usage patterns and statistics
  - Most used profiles, regions, instances
  - Connection success/failure rates
  - Optional telemetry with privacy controls
- [ ] **Structured logging**: Add logging with different verbosity levels
  - Debug, info, warn, error levels
  - Optional log file output
- [ ] **Health checks**: Verify SSH connectivity before attempting connection
  - Test SSH port availability
  - Validate SSH key authentication
- [ ] **Connection time**: Show connection establishment time metrics

### Security & Compliance
- [ ] **Session recording**: Optional session logging for compliance
  - Configurable session recording
  - Integration with audit systems
- [ ] **MFA support**: Better integration with AWS MFA workflows
  - Support for hardware tokens
  - Automatic MFA token refresh
- [ ] **Key management**: Integration with SSH agent or key management
  - Automatic SSH key discovery
  - Support for encrypted SSH keys
- [ ] **Audit trail**: Log all connections with timestamps and user info

### Advanced Features
- [ ] **Bulk operations**: Run commands on multiple instances simultaneously
  - Execute same command across selected instances
  - Parallel execution with progress tracking
- [ ] **File transfer**: Built-in SCP/SFTP support for file transfers
  - Drag-and-drop file transfer interface
  - Progress bars for large transfers
- [ ] **Port forwarding**: Easy SSH tunnel management
  - Configure and manage SSH tunnels
  - Support for local and remote port forwarding
- [ ] **Session management**: Named sessions and session sharing
  - Save and restore SSH session configurations
  - Share session configs with team members

## 🌐 Future Enhancements

### Multi-cloud & Integration
- [ ] **Other cloud providers**: Support for GCP, Azure instances
  - Pluggable provider architecture
  - Unified interface across cloud providers
- [ ] **Kubernetes**: Integration with kubectl for container access
  - Discover and connect to Kubernetes pods
  - Support for kubectl exec and port-forward
- [ ] **Docker**: Support for connecting to Docker containers
  - Local and remote Docker daemon support
  - Container discovery and connection
- [ ] **Terraform**: Integration with Terraform state for instance discovery
  - Read Terraform state files
  - Discover instances from Terraform plans

### Modern Features
- [ ] **TUI improvements**: Better keyboard shortcuts and mouse support
  - Vim-style key bindings
  - Mouse support for selection and navigation
- [ ] **Plugins**: Plugin system for extending functionality
  - Go plugin architecture
  - Community plugin repository
- [ ] **JSON/YAML output**: Machine-readable output formats
  - Support for structured data output
  - Integration with other tools and scripts
- [ ] **REST API**: Optional HTTP API for programmatic access
  - RESTful API for instance discovery and connection
  - Authentication and authorization
- [ ] **WebUI**: Optional web interface for team sharing
  - Browser-based interface
  - Multi-user support with permissions

### Distribution & Packaging
- [ ] **Docker image**: Containerized version for consistent environments
  - Multi-architecture Docker images
  - Integration with container orchestration
- [ ] **Binary releases**: Automated releases with GitHub Actions
  - Cross-platform binary builds
  - Automated testing and release pipeline
- [ ] **Package managers**: APT, RPM, Chocolatey support
  - Native package manager integration
  - Automatic dependency management

---

## 🎯 Next Steps

1. **Phase 1**: Implement caching, custom SSH args, and connection history
2. **Phase 2**: Add Homebrew formula and improve search capabilities
3. **Phase 3**: Focus on configuration customization and advanced features
4. **Phase 4**: Explore multi-cloud and modern feature additions

## 💡 Ideas for Future Consideration

- Integration with password managers (1Password, Bitwarden)
- Support for SSH certificate authentication
- Integration with cloud cost management tools
- Mobile app for basic instance management
- Integration with monitoring tools (DataDog, New Relic)
- Support for custom instance provisioning workflows
- Integration with CI/CD pipelines for deployment workflows