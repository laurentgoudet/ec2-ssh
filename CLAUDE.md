# Build Instructions

## Correct Build Command
Use the build script to build the application:
```bash
./build.sh
```

This builds the application from `./cmd/ec2-ssh` and outputs to `ec2-ssh` executable.

## Testing Commands
After building, test with:
```bash
./ec2-ssh --version
./ec2-ssh -v
```

## Do NOT use
- `go build` (builds wrong target)
- `go build .` (builds wrong target)

Always use `./build.sh` for correct builds.