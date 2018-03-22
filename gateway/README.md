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
    * Use query parameter `access_token` to carry authorization token

- TCP long connection gateway
    * Required authorization
    * Use the first packet of the connection to carry authorization information:<br>Package type `PULL`, URI `/auth/verify`, BodyType `s`, Body `access token string`
