# stethoscope

Listen to your network

## Usage

```
./stethoscope help
NAME:
   stethoscope - listen to your network

USAGE:
   main [global options] command [command options] [arguments...]

COMMANDS:
   net      options for network operations
   version  print version information
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --version, -v  print version (default: false)
   --help, -h     show help (default: false)
```

```
./stethoscope net help
NAME:
   stethoscope net - options for network operations

USAGE:
   stethoscope net [global options] command [command options] [arguments...]

COMMANDS:
   list     print network interface list
   stat     print network interface statistics
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

```
./stethoscope net help list
NAME:
   stethoscope net list - print network interface list

USAGE:
   stethoscope net list [arguments...]
```

```
./stethoscope net help stat
NAME:
   stethoscope net stat - print network interface statistics

USAGE:
   stethoscope net stat [command options] [arguments...]

OPTIONS:
   --filter value, -f value     capture filter
   --interface value, -i value  capture interface
```

## Examples

```
# List all network interfaces
./stethoscope net list

# List local loopback network interface
./stethoscope net list lo

# Stat TCP connections on Loopback
./stethoscope net stat -i eth0 -f "tcp port 80" -f "tcp port 443"

# Stat TCP connections to google.com
./stethoscope net stat -i eth0 -f "host google.com"
```