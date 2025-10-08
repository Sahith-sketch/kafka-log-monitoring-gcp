package models

import "time"

type AuditLog struct {
	Timestamp       time.Time              `json:"timestamp"`
	ReceiveTimestamp time.Time             `json:"receiveTimestamp,omitempty"`
	Severity        string                 `json:"severity"`
	LogName         string                 `json:"logName"`
	InsertId        string                 `json:"insertId,omitempty"`
	Resource        Resource               `json:"resource"`
	ProtoPayload    *Payload               `json:"protoPayload,omitempty"`
	HttpRequest     *HttpRequest           `json:"httpRequest,omitempty"`
	Operation       *Operation             `json:"operation,omitempty"`
	Trace           string                 `json:"trace,omitempty"`
	SpanId          string                 `json:"spanId,omitempty"`
	SourceLocation  *SourceLocation        `json:"sourceLocation,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	TextPayload     string                 `json:"textPayload,omitempty"`
	JsonPayload     map[string]interface{} `json:"jsonPayload,omitempty"`
}

type Resource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels,omitempty"`
}

type Payload struct {
	Type                string                 `json:"@type"`
	MethodName          string                 `json:"methodName,omitempty"`
	ResourceName        string                 `json:"resourceName,omitempty"`
	ServiceName         string                 `json:"serviceName,omitempty"`
	CallerIp            string                 `json:"callerIp,omitempty"`
	CallerSuppliedUserAgent string             `json:"callerSuppliedUserAgent,omitempty"`
	NumResponseItems    int64                  `json:"numResponseItems,omitempty"`
	Status              *Status                `json:"status,omitempty"`
	AuthenticationInfo  *AuthenticationInfo    `json:"authenticationInfo,omitempty"`
	AuthorizationInfo   []AuthorizationInfo    `json:"authorizationInfo,omitempty"`
	RequestMetadata     *RequestMetadata       `json:"requestMetadata,omitempty"`
	Request             map[string]interface{} `json:"request,omitempty"`
	Response            map[string]interface{} `json:"response,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

type HttpRequest struct {
	RequestMethod                  string `json:"requestMethod,omitempty"`
	RequestUrl                     string `json:"requestUrl,omitempty"`
	RequestSize                    string `json:"requestSize,omitempty"`
	Status                         int    `json:"status,omitempty"`
	ResponseSize                   string `json:"responseSize,omitempty"`
	UserAgent                      string `json:"userAgent,omitempty"`
	RemoteIp                       string `json:"remoteIp,omitempty"`
	ServerIp                       string `json:"serverIp,omitempty"`
	Referer                        string `json:"referer,omitempty"`
	Latency                        string `json:"latency,omitempty"`
	CacheLookup                    bool   `json:"cacheLookup,omitempty"`
	CacheHit                       bool   `json:"cacheHit,omitempty"`
	CacheValidatedWithOriginServer bool   `json:"cacheValidatedWithOriginServer,omitempty"`
	CacheFillBytes                 string `json:"cacheFillBytes,omitempty"`
	Protocol                       string `json:"protocol,omitempty"`
}

type Operation struct {
	Id       string `json:"id,omitempty"`
	Producer string `json:"producer,omitempty"`
	First    bool   `json:"first,omitempty"`
	Last     bool   `json:"last,omitempty"`
}

type SourceLocation struct {
	File     string `json:"file,omitempty"`
	Line     string `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

type Status struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type AuthenticationInfo struct {
	PrincipalEmail string `json:"principalEmail,omitempty"`
}

type AuthorizationInfo struct {
	Resource   string `json:"resource,omitempty"`
	Permission string `json:"permission,omitempty"`
	Granted    bool   `json:"granted,omitempty"`
}

type RequestMetadata struct {
	CallerIp                string `json:"callerIp,omitempty"`
	CallerSuppliedUserAgent string `json:"callerSuppliedUserAgent,omitempty"`
	CallerNetwork           string `json:"callerNetwork,omitempty"`
	RequestAttributes       map[string]interface{} `json:"requestAttributes,omitempty"`
	DestinationAttributes   map[string]interface{} `json:"destinationAttributes,omitempty"`
}

type IntrusionAlert struct {
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	SourceIP    string    `json:"sourceIP"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
}
