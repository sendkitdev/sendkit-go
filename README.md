# SendKit Go SDK

Official Go SDK for the [SendKit](https://sendkit.com) email API.

## Installation

```bash
go get github.com/sendkitdev/sendkit-go
```

## Usage

### Create a Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    sendkit "github.com/sendkitdev/sendkit-go"
)

func main() {
    client, err := sendkit.NewClient("sk_your_api_key")
    if err != nil {
        log.Fatal(err)
    }

    // Send an email
    resp, err := client.Emails.Send(context.Background(), &sendkit.SendEmailParams{
        From:    "you@example.com",
        To:      []string{"recipient@example.com"},
        Subject: "Hello from SendKit",
        HTML:    "<h1>Welcome!</h1>",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.ID)
}
```

### Send a MIME Email

```go
resp, err := client.Emails.SendMime(context.Background(), &sendkit.SendMimeEmailParams{
    EnvelopeFrom: "you@example.com",
    EnvelopeTo:   "recipient@example.com",
    RawMessage:   mimeString,
})
```

### Error Handling

API errors are returned as `*sendkit.APIError`:

```go
resp, err := client.Emails.Send(ctx, params)
if err != nil {
    var apiErr *sendkit.APIError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.Name)       // e.g. "validation_error"
        fmt.Println(apiErr.Message)    // e.g. "The to field is required."
        fmt.Println(apiErr.StatusCode) // e.g. 422
    }
}
```

### Configuration

```go
// Read API key from SENDKIT_API_KEY environment variable
client, _ := sendkit.NewClient("")

// Custom base URL
client, _ := sendkit.NewClient("sk_...", sendkit.WithBaseURL("https://custom.api.com"))

// Custom HTTP client
client, _ := sendkit.NewClient("sk_...", sendkit.WithHTTPClient(&http.Client{
    Timeout: 30 * time.Second,
}))
```
