# protoc-gen-go-http

`protoc-gen-go-http` 读取 proto service 上的 `google.api.http` 注解，并生成适配 `github.com/nexuer/corex/transport/http` 的 HTTP 服务端和客户端代码。

生成器的 HTTP 映射行为整体保持接近 Kratos，但路由规则刻意收紧为 Go `net/http.ServeMux` 兼容形式。当前服务端注册仍然面向 gorilla/mux 运行时，因此会在生成阶段把 ServeMux wildcard 转成 gorilla/mux 可注册的形式。

English documentation: [README.md](./README.md)

## 路由规则

`get`、`post`、`put`、`delete`、`patch` 和 `custom.path` 的值必须是纯 path pattern。不要在 path 字符串里写 HTTP method 或 host；HTTP method 已经由 `google.api.http` 的字段表达。

支持的 path 形式：

```proto
option (google.api.http) = {
  get: "/v1/users"
};
```

静态路径，不绑定路径变量。

```proto
option (google.api.http) = {
  get: "/v1/users/{id}"
};
```

单段 wildcard。wildcard 名称必须匹配请求消息的顶层字段。

```proto
message GetUserRequest {
  string id = 1;
}
```

```proto
option (google.api.http) = {
  get: "/v1/orgs/{org_id}/users/{user_id}"
};
```

多个单段 wildcard。每个 wildcard 都映射到请求消息的顶层字段。

```proto
option (google.api.http) = {
  get: "/v1/files/{path...}"
};
```

尾部 catch-all wildcard。它必须是最后一个 path segment。

```proto
option (google.api.http) = {
  get: "/v1/users/{$}"
};
```

结束锚点。这里使用 ServeMux 的 `{$}` 形式，让路径精确匹配，而不是匹配整个子树。

## 路由限制

路由会先用 Go `net/http.ServeMux` pattern 规则校验，再转换给当前 gorilla/mux 运行时注册。

允许的 wildcard 形式：

- `{name}` 匹配一个 path segment
- `{name...}` 匹配剩余 path
- `{$}` 表示 ServeMux 结束锚点

拒绝的例子：

```proto
get: "/v1/{name=messages/*}"  // Google transcoding pattern
get: "/v1/{message.id}"       // 嵌套字段
get: "/v1/{id}.json"          // wildcard 不是完整 segment
get: "/v1/{path...}/tail"     // catch-all 不是最后一段
get: "GET /v1/{id}"           // method 不能写进 path
get: "example.com/v1/{id}"    // host 不能写进 path
```

wildcard 名称必须是合法 Go identifier，例如 `id`、`user_id`、`path`。

注册到 gorilla/mux 时：

- `{name}` 保持为 `{name}`
- `{name...}` 转成 `{name:.*}`
- `/{$}` 注册成 `/`
- `/prefix/{$}` 注册成 `/prefix/`

## HTTP 方法

支持的 method pattern：

```proto
get: "/v1/users/{id}"
put: "/v1/users/{id}"
post: "/v1/users"
delete: "/v1/users/{id}"
patch: "/v1/users/{id}"
custom {
  kind: "REPORT"
  path: "/v1/report"
}
```

`custom` 会保留，行为和 Kratos 一致，但需要谨慎使用，因为代理、浏览器、网关和文档工具对自定义 HTTP method 的支持并不总是一致。

## Prefix

`-prefix` 选项会给每个带注解的路由添加统一 path 前缀，并在拼接后再做路由校验：

```yaml
plugins:
  - local: protoc-gen-go-http
    out: api
    opt:
      - paths=source_relative
      - prefix=/api/console/v1
```

前缀和 path 连接处的斜杠会被规范化：

```text
prefix=/api/
path=/users/{id}
=> /api/users/{id}
```

生成器不会清理或重写 prefix/path 内部内容。最终拼接出来的路由必须能通过 ServeMux 规则。

## Body 规则

无 body：

```proto
option (google.api.http) = {
  get: "/v1/users/{id}"
};
```

生成的服务端会把 query 参数和 path 变量绑定到请求消息。

整个请求作为 body：

```proto
option (google.api.http) = {
  post: "/v1/users"
  body: "*"
};
```

生成的服务端会把整个 HTTP body 绑定到请求消息。

指定 body 字段：

```proto
message UpdateUserRequest {
  string id = 1;
  User user = 2;
}

option (google.api.http) = {
  patch: "/v1/users/{id}"
  body: "user"
};
```

只有 `user` 字段从 HTTP body 解码。path 和 query 参数仍然会绑定到完整请求消息。

## Response Body

```proto
message GetUserResponse {
  User user = 1;
}

option (google.api.http) = {
  get: "/v1/users/{id}"
  response_body: "user"
};
```

生成的服务端只会把 `response.user` 写入 HTTP response body。

## Additional Bindings

```proto
option (google.api.http) = {
  get: "/v1/users/{id}"
  additional_bindings {
    get: "/v1/me"
  }
};
```

每个 binding 都会为同一个 RPC 生成一条独立的 HTTP handler 注册。

## Streaming 映射

Unary RPC 会生成普通 HTTP handler 和 client。

Server-streaming RPC 映射为 Server-Sent Events。

Client-streaming 和 bidirectional-streaming RPC 映射为 WebSocket。

生成代码需要 `corex/transport/http` 提供匹配的 stream 支持。

## 生成内容

对每个带注解的 service，生成器会输出：

- operation 常量
- `XxxHTTPServer` 接口
- `RegisterXxxHTTPServer`
- 每个 binding 对应的 HTTP handler
- 必要时生成 stream server wrapper
- `XxxHTTPClient` 接口
- HTTP client 实现
- 必要时生成 stream client wrapper

没有 service 或没有 `google.api.http` 注解的文件会被跳过。
