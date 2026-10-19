package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed template.tmpl
var httpTemplate string

type serviceDesc struct {
	// ServiceType 是 Go 侧 service 类型名，例如 Greeter，用于生成 GreeterHTTPServer。
	ServiceType string
	// ServiceName 是 proto service 全名，例如 helloworld.v1.Greeter，用于 operation 名称。
	ServiceName string
	// Metadata 是 proto 文件路径，保留给模板或后续生成元信息使用。
	Metadata string
	// Methods 是所有 HTTP binding 展开后的方法列表；additional_bindings 会生成多条记录。
	Methods []*methodDesc
	// MethodSets 按 Go 方法名去重，用于生成 service/client 接口和 operation 常量。
	MethodSets map[string]*methodDesc
}

type methodDesc struct {
	// Name 是 Go 方法名，例如 SayHello。
	Name string
	// OriginalName 是 proto 原始 rpc 名称，用于生成 operation 常量。
	OriginalName string
	// Num 是同名方法的 binding 序号，避免 additional_bindings 生成的 handler 重名。
	Num int
	// Request 是请求 message 的 Go 类型名。
	Request string
	// Reply 是响应 message 的 Go 类型名。
	Reply string
	// Comment 是 rpc 注释和 deprecated 注释，直接写入生成的接口方法上。
	Comment string

	// Path 是服务端注册用路径。
	Path string
	// PathTemplate 是客户端构造请求路径用的原始模板，保留 {id} / {path...} 语义。
	PathTemplate string
	// Method 是 HTTP 方法，例如 GET、POST、PATCH；custom rule 会使用自定义 kind。
	Method string
	// HasVars 表示路径中是否包含 {id}、{path...} 等路径变量。
	HasVars bool
	// HasBody 表示 google.api.http 是否声明了 body。
	HasBody bool
	// Body 是生成代码访问请求 body 字段的 Go 字段路径，例如 ".User"；body:"*" 时为空。
	Body string
	// BodyField 是 proto 里的 body 字段名，例如 "user"；body:"*" 时为 "*"。
	BodyField string
	// BodyQueryName 是 body 字段的 JSON 名称，用于客户端构造 query 时排除该字段。
	BodyQueryName string
	// BodyHTTPBody 表示请求 body 或命名 body 字段是否为 google.api.HttpBody。
	BodyHTTPBody bool
	// BodyMessage 表示命名 body 字段是否是非 repeated/map 的 message；客户端流可逐帧发送该字段。
	BodyMessage bool
	// ResponseBody 是生成代码访问响应 body 字段的 Go 字段路径，例如 ".Data"。
	ResponseBody string
	// ResponseBodyHTTPBody 表示 response_body 指向的字段是否为 google.api.HttpBody。
	ResponseBodyHTTPBody bool
	// ReplyHTTPBody 表示整个响应 message 是否为 google.api.HttpBody。
	ReplyHTTPBody bool
	// ClientStreaming 表示 rpc 是否为客户端流；当前模板映射到 WebSocket。
	ClientStreaming bool
	// ServerStreaming 表示 rpc 是否为服务端流；当前模板映射到 SSE，双向流映射到 WebSocket。
	ServerStreaming bool
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
