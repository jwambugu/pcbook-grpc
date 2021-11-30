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

## Running Server

To run the client, run the following command

```bash
  make run-client
```

To run the server, run the following command

```bash
  make run-server
```

## Running Tests

To run tests, run the following command

```bash
  make test
```