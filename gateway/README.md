# Gateway

Package gateway is the main program for TCP and HTTP services.

## Demo

```go
package main

import (
    "github.com/henrylee2cn/cfgo"
    "github.com/xiaoenai/ants/gateway"
)

func main() {
    cfg := gateway.NewConfig()
    cfgo.MustReg("gateway", cfg)
    // Run a gateway instance with default business logic and default socket protocol.
    gateway.Run(*cfg, nil, nil)
}
```

## Usage

### Authorization

- HTTP short connection gateway
    * Optional authorization
    * Use query or header parameter to carry authorization token

- TCP long connection gateway
    * Required authorization
    * Use the first packet of the connection to carry authorization information:<br>Package type `PULL`, URI `/auth/verify`, BodyType `s`, Body `access token string`

### RequestID

- HTTP short connection gateway
    * Optional query parameter
    * Use query parameter `_seq` to carry request ID

- TCP long connection gateway
    * Required packet `seq` field
    * The request ID is `{session ID}@{packet seq}`

### HTTP Status Code

- 200 OK
- 299 Business error
- 500 Internal communication error
- Other codes (200,600)
