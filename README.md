# stethoscope

Listen to your network

## Usage

```
./steth help
NAME:
   steth - listen to your network

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
./steth net help
NAME:
   steth net - options for network operations

USAGE:
   steth net [global options] command [command options] [arguments...]

COMMANDS:
   list     print network interface list
   stat     print network interface statistics
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

```
./steth net help list
NAME:
   steth net list - print network interface list

USAGE:
   steth net list [arguments...]
```

```
./steth net help stat
NAME:
   steth net stat - print network interface statistics

USAGE:
   steth net stat [command options] [arguments...]

OPTIONS:
   --filter value, -f value     capture filter
   --interface value, -i value  capture interface
```

## Examples

```
# List all network interfaces
./steth net list

# List local loopback network interface
./steth net list lo

# Stat TCP connections on Loopback
./steth net stat -i eth0 -f "tcp port 80" -f "tcp port 443"

# Stat TCP connections to google.com
./steth net stat -i eth0 -f "host google.com"
```