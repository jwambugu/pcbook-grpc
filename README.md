# [wip] pcbook

This is web service that will allow us to manage and search for laptop configurations

## Services

The web service has the following services:

1. LaptopService

| RPC            | REQUEST TYPE          | RESPONSE TYPE          | DESCRIPTION                                       |
| :---           | :---                  |  :---                  | :---                                              |
| CreateLaptop   | CreateLaptopRequest   | CreateLaptopResponse   | Creates and stores a new laptop                   |
| SearchLaptop   | SearchLaptopRequest   | SearchLaptopResponse   |  Searches for a laptop using the provided `Filter`|
| UploadImage    | UploadImageRequest    | UploadImageResponse    |  Uploads and stores a laptop image                |
| RateLaptop     | RateLaptopRequest     | RateLaptopResponse     |  Rates a laptop                                   |

## Generate TLS Certificates

To run the client and the server on `tls mode`, you need to generate the certificates.

```bash
  make gen-cert
```

## Running Servers

By default, all servers run with the `enable-tls` flag enabled.

To run the client, run the following command:

```bash
  make run-client
```

The app supports two types of servers:

- gRPC
- REST (HTTP) using [GRPC Gateway](https://github.com/grpc-ecosystem/grpc-gateway)

To run the gRPC server, run the following command:

```bash
  make run-grpc-server
```

To run the REST server, run the following command:

```bash
  make run-rest-server
```

## Running Tests

To run tests, run the following command

```bash
  make test
```
