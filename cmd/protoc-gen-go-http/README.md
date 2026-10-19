# protoc-gen-go-http

`protoc-gen-go-http` reads `google.api.http` annotations from proto services and generates HTTP server/client adapters for `github.com/nexuer/corex/transport/http`.

The generator keeps the HTTP mapping behavior close to Kratos, while intentionally restricting route patterns to a Go `net/http.ServeMux` compatible form. Server registration still targets the current gorilla/mux-based runtime, so ServeMux wildcards are translated where needed.

Chinese documentation: [README.zh-CN.md](./README.zh-CN.md)

## Routing Rules

The `get`, `post`, `put`, `delete`, `patch`, and `custom.path` values must be path-only patterns. Do not include a method or host in the path string; the HTTP method is already represented by the `google.api.http` field.

Supported path forms:

```proto
option (google.api.http) = {
  get: "/v1/users"
};
```

Static path. No path variables are bound.

```proto
option (google.api.http) = {
  get: "/v1/users/{id}"
};
```

Single-segment wildcard. The wildcard name must match a top-level request field.

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

Multiple single-segment wildcards. Each wildcard maps to a top-level request field.

```proto
option (google.api.http) = {
  get: "/v1/files/{path...}"
};
```

Trailing catch-all wildcard. It must be the final path segment.

```proto
option (google.api.http) = {
  get: "/v1/users/{$}"
};
```

End anchor. This uses the ServeMux `{$}` form to match the path exactly instead of a subtree.

## Route Restrictions

Routes are validated with Go `net/http.ServeMux` pattern rules, then converted for the current gorilla/mux runtime.

Accepted wildcard forms:

- `{name}` for one path segment
- `{name...}` for the remaining path
- `{$}` as the ServeMux end anchor

Rejected examples:

```proto
get: "/v1/{name=messages/*}"  // Google transcoding pattern
get: "/v1/{message.id}"       // nested field
get: "/v1/{id}.json"          // wildcard is not a full segment
get: "/v1/{path...}/tail"     // catch-all is not final
get: "GET /v1/{id}"           // method must not be part of path
get: "example.com/v1/{id}"    // host must not be part of path
```

Wildcard names must be valid Go identifiers, such as `id`, `user_id`, or `path`.

For gorilla/mux registration:

- `{name}` stays `{name}`
- `{name...}` becomes `{name:.*}`
- `/{$}` is registered as `/`
- `/prefix/{$}` is registered as `/prefix/`

## HTTP Methods

Supported method patterns:

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

`custom` is supported, matching Kratos behavior, but should be used carefully because proxies, browsers, gateways, and documentation tooling may not handle custom HTTP methods consistently.

## Prefix

The `-prefix` option prepends a common path prefix to every annotated route before route validation:

```yaml
plugins:
  - local: protoc-gen-go-http
    out: api
    opt:
      - paths=source_relative
      - prefix=/api/console/v1
```

Boundary slashes are normalized:

```text
prefix=/api/
path=/users/{id}
=> /api/users/{id}
```

The generator does not clean or rewrite the inside of either value. The final joined route must be accepted by ServeMux rules.

## Body Rules

No body:

```proto
option (google.api.http) = {
  get: "/v1/users/{id}"
};
```

The generated server binds query parameters and path variables into the request.

Whole request body:

```proto
option (google.api.http) = {
  post: "/v1/users"
  body: "*"
};
```

The generated server binds the whole HTTP body into the request message.

Named body field:

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

Only the `user` field is decoded from the HTTP body. Path and query parameters are still bound to the full request.

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

The generated server writes only `response.user` as the HTTP response body.

## Additional Bindings

```proto
option (google.api.http) = {
  get: "/v1/users/{id}"
  additional_bindings {
    get: "/v1/me"
  }
};
```

Each binding generates a separate HTTP handler registration for the same RPC.

## Streaming Mapping

Unary RPCs generate ordinary HTTP handlers and clients.

Server-streaming RPCs are mapped to Server-Sent Events.

Client-streaming and bidirectional-streaming RPCs are mapped to WebSocket.

The generated code requires matching stream support in `corex/transport/http`.

## Generation Summary

For each annotated service, the generator emits:

- operation constants
- `XxxHTTPServer` interface
- `RegisterXxxHTTPServer`
- HTTP handlers for each binding
- stream server wrappers when needed
- `XxxHTTPClient` interface
- HTTP client implementation
- stream client wrappers when needed

Files without services or without `google.api.http` annotations are skipped.
