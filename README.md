
# README

## Introduction
This is a Go client for the OpenAI API. It provides a simple interface for accessing the OpenAI API, including functions for uploading files, fine-tuning models, and moderating text. To use this client, you will need an OpenAI API key. You can obtain an API key by signing up for the OpenAI API at https://beta.openai.com/signup/. Once you have an API key, you can use this client to access the OpenAI API and build powerful AI applications.

# OpenAI Golang SDK

This is a Golang SDK for the OpenAI API. It provides a simple interface for accessing the various OpenAI API endpoints.

## Installation

To install the SDK, simply run:

```sh
go get github.com/neoguojing/openai
```

## Usage

To use the SDK, you will need an API key from OpenAI. You can obtain an API key by signing up for an account on the OpenAI website.

Once you have an API key, you can create a new instance of the `OpenAI` struct:

```go
import "github.com/neoguojing/openai"

openai := openai.NewOpenAI("your-api-key")
```

You can then use the various methods provided by the SDK to interact with the OpenAI API.

## Examples

Here are some examples of how to use the SDK:

```go
// Retrieve a list of available models
modelList, err := openai.Model().List()
if err != nil {
    log.Fatal(err)
}
fmt.Println(modelList)

// Generate completions for a given prompt
completions, err := openai.Completions("Hello, world!", 5)
if err != nil {
    log.Fatal(err)
}
fmt.Println(completions)

// Generate an image from a given prompt
image, err := openai.Image().Generate("A cute cat", 1, "512x512")
if err != nil {
    log.Fatal(err)
}
fmt.Println(image)
```

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue on the GitHub repository.

## License

This SDK is licensed under the Apache License.

