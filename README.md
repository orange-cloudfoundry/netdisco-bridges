# Netdisco-bridges

This is a suite of bridges for [netdisco](https://github.com/netdisco/netdisco). With this you will be able:

- find devices from criteria with dns queries (with this you can plug [prometheus](https://prometheus.io/) for monitoring)
- Get reports and device information on openmetrics format (usable on [prometheus](https://prometheus.io/))
- Use an api for
    - getting suite of devices (entries) in list or by domain
    - getting routes associated to entries for setting access to other system

## Getting started

1. Download latest release for your system
2. Create a `config.yml` file with this content for a dev deployment

```yaml
log:
  level: debug
dns_server:
  listen: 127.0.0.1:8853
http_server:
  listen: 127.0.0.1:8080
netdisco:
  endpoint: https://my.netdisco.com
  username: user
  password: 'password'

entries:
  - domain: all.netdisco
    routing:
      scheme: https
      host: '{{ .IP }}'
      metadata:
        entryPoints: [ https ]
    targets:
      - q: '%'
```

3. run with `./netdisco-bridges --config config.yml`

## Usable Bridges

### As DNS

- `dig @127.0.0.1 -p 8853 all.netdisco` - Gave all IPs set for entries
- `dig @127.0.0.1 -p 8853 all.netdisco SRV` - Gave all dns set for entries
- `dig @127.0.0.1 -p 8853 all.netdisco TXT` - Gave all devices information in base64 json encoded set for entries

### With API

- `http://127.0.0.1:8080/api/v1/entries/*/routes` - Gave all http routes formatted for all entries
- `http://127.0.0.1:8080/api/v1/entries/*/routes/traefik` - Gave all http routes formatted for all entries in traefik format for using as provider
- `http://127.0.0.1:8080/api/v1/entries/{domain entry}/routes?format=default` - Gave http routes formatted for specified entry
- `http://127.0.0.1:8080/api/v1/entries` - List all entries set
- `http://127.0.0.1:8080/api/v1/entries/{domain}/devices` - Gave all devices formatted for specified entry

## Configuration

For understanding config definition format:

- `[]` means optional (by default parameter is required)
- `<>` means type to use

### Root configuration in config.yml

```yaml
dns_server:
  # set to true to disable dns server
  [ disabled: <bool> ]
  # Listen address for listening for dns
  [ listen: <string> | default = 0.0.0.0:53 ]

http_server:
  # set to true to disable http server
  [ disabled: <bool> ]
  # Listen address for listening for http
  [ listen: <string> | default = 0.0.0.0:8080 or 0.0.0.0:8443 if ssl enabled ]
  # set to true to enable tls
  [ enable_ssl: <bool> ]
  tls_pem:
    # cert chain in pem format when tls enabled
    [ cert_chain: <string> ]
    # private key in pem format when tls enabled
    [ private_key: <string> ]

log:
  # log level to use for server
  # you can chose: `trace`, `debug`, `info`, `warn`, `error`, `fatal` or `panic`
  [ level: <string> | default = info ]
  # Set to true to force not have color when seeing logs
  [ no_color: <bool> ]
  # et to true to see logs as json format
  [ in_json: <bool> ]

netdisco:
  # url pointing to your netdisco
  endpoint: <string>
  # Username for connecting to netdisco
  username: <string>
  # Password for connecting to netdisco
  password: <string>
  # set to true to not verify ssl certificate
  [ insecure_skip_verify: <bool> ]

# Netdisco-bridges load devices set in entries async for performance and caching purpose over netdisco
# you can change workers profile here
workers:
  # number of workers to use for loading entries
  # Set more than entries is useless
  [ nb_workers: <int> | default = 5 ]
  # Interval for data to be refreshed from netdisco
  [ refresh_interval: <duration> | default = "25m" ]

# list of entry (defined below)
entries:
- <entry>
```

### entry configuration

```yaml
# Domain will be used for dns/http api for getting list of devices associated
domain: <string>
# Netdisco search criteria, at least one is required
targets:
  # Partial match of Device contact, serial, chassis ID, module serials, location, name, description, dns, or any IP alias
  # % can give all device
  [ q: <string> ]
  # Partial match of the Device name
  [ name: <string> ]
  # Partial match of the Device location
  [ location: <string> ]
  # Partial match of any of the Device IP aliases
  [ dns: <string> ]
  # IP or IP Prefix within which the Device must have an interface address
  [ ip: <string> ]
  # Partial match of the Device description
  [ description: <string> ]
  # MAC Address of the Device or any of its Interfaces
  [ mac: <string> ]
  # Exact match of the Device model
  [ model: <string> ]
  # Exact match of the Device operating system
  [ os: <string> ]
  # Exact match of the Device operating system version
  [ os_ver: <string> ]
  # Exact match of the Device vendor
  [ vendor: <string> ]
  # OSI Layer which the device must support
  [ layers: <string> ]
  # If true, all fields (except “q”) must match the Device
  [ matchall: <bool> ]
# routing let create an http route based on criteria for each device found in entry
# if not set no route will be associated to this set of devices
# config defined below
[ routing: <routing> ]
```

### routing configuration

Templating is allowed here, you have access to all function defined here: https://masterminds.github.io/sprig/

Device information become accessible for each value, device has those informations:

- `UptimeAge`
- `Location`
- `SinceLastArpnip`
- `FirstSeenStamp`
- `OsVer`
- `Name`
- `LastArpnipStamp`
- `Model`
- `SinceFirstSeen`
- `IP`
- `Serial`
- `SinceLastMacsuck`
- `DNS`
- `SinceLastDiscover`
- `LastMacsuckStamp`
- `LastDiscoverStamp`

```yaml
# Scheme to use to create route
[ scheme: <string|template> | default = "https" ]
# Port for accessing to route
[ port: <string|template> | default = not set ]
# Host to set for the route
[ host: <string|template> | default = not set ]
# metadata for let formatter do its magic
# for now, only traefik use it
# you can set `entryPoints` for traefik and `enableTls` to true to enable resolve on traefik on tls also.
metadata:
  <string|template>: <map|template>
```
