## khronos

golang 优先级队列，兼容 redis 协议但不依赖 redis

### Example

```go

package main

import (
    "fmt"
    "github.com/eatmoreapple/khronos"
)


func main() {
    khronos.ListenAndServe(":7464")
}

```

### Commands

```shell
redis-cli -p 7464

> ping
PONG

> ping hello
"hello"

> push queue1 mydata 1
OK

> pop queue1 
"mydata"

> push queue1 mydata2 1
OK

> push queue1 mydata3 3
OK

> pop queue1
"mydata3"
```

