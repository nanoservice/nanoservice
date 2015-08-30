# nanoservice CLI tool

This is part of nanoservice core.

## Installation

To build it yourself, make sure you have recent version of Go setup and run:

```bash
go get github.com/nanoservice/nanoservice
```

In the future this will be possible:

```bash
# TODO: make this installer and build static binaries
curl -L https://github.com/nanoservice/installer/raw/stable/install.sh | bash
```

## Usage

### Configure cluster on AWS (default option)

```bash
nanoservice configure
# or
nanoservice configure --aws
```

### Configure hosted cluster

```bash
nanoservice configure --hosted
```

### Configure local docker cluster (for development)

You need to have proper docker setup locally (or remote docker server with configured docker client).

```bash
nanoservice configure --docker
```

### Create a nanoservice

```bash
nanoservice create --TEMPLATE --LANGUAGE NAME
```

Example: `nanoservice --web --golang helloworld`.

### Deploy a nanoservice

```bash
nanoservice deploy
```

### Scale a nanoservice

```bash
nanoservice scale 3

# or to turn it off:
nanoservice scale 0
```

## Development

Use normal TDD development style.

## Contributing

1. Fork it ( https://github.com/nanoservice/nanoservice/fork )
1. Create your feature branch (git checkout -b my-new-feature)
1. Commit your changes (git commit -am 'Add some feature')
1. Push to the branch (git push origin my-new-feature)
1. Create a new Pull Request

## Contributors

* [waterlink](https://github.com/waterlink) Oleksii Fedorov, creator, maintainer
