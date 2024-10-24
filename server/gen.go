//go:build go1.22

// Package sentinel provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package sentinel

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/runtime"
)

// Defines values for Decision.
const (
	Approve   Decision = "approve"
	Escalate  Decision = "escalate"
	Modify    Decision = "modify"
	Reject    Decision = "reject"
	Terminate Decision = "terminate"
)

// Defines values for Status.
const (
	Completed  Status = "completed"
	Processing Status = "processing"
	Queued     Status = "queued"
	Timeout    Status = "timeout"
)

// Arguments defines model for Arguments.
type Arguments struct {
	Cmd  *string `json:"cmd,omitempty"`
	Code *string `json:"code,omitempty"`
}

// AssistantMessage defines model for AssistantMessage.
type AssistantMessage struct {
	Content   string      `json:"content"`
	Role      string      `json:"role"`
	Source    *string     `json:"source,omitempty"`
	ToolCalls *[]ToolCall `json:"tool_calls,omitempty"`
}

// Choice defines model for Choice.
type Choice struct {
	Message    AssistantMessage `json:"message"`
	StopReason *string          `json:"stop_reason,omitempty"`
}

// CodeSnippet defines model for CodeSnippet.
type CodeSnippet struct {
	Text string `json:"text"`
}

// Decision defines model for Decision.
type Decision string

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	Status string `json:"status"`
}

// HubStats defines model for HubStats.
type HubStats struct {
	AssignedReviews    map[string]int `json:"assigned_reviews"`
	BusyClients        int            `json:"busy_clients"`
	CompletedReviews   int            `json:"completed_reviews"`
	ConnectedClients   int            `json:"connected_clients"`
	FreeClients        int            `json:"free_clients"`
	QueuedReviews      int            `json:"queued_reviews"`
	ReviewDistribution map[string]int `json:"review_distribution"`
	StoredReviews      int            `json:"stored_reviews"`
}

// LLMExplanation defines model for LLMExplanation.
type LLMExplanation struct {
	Explanation string  `json:"explanation"`
	Score       float32 `json:"score"`
}

// LLMPrompt defines model for LLMPrompt.
type LLMPrompt struct {
	Prompt string `json:"prompt"`
}

// LLMPromptResponse defines model for LLMPromptResponse.
type LLMPromptResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Message defines model for Message.
type Message struct {
	Content    string      `json:"content"`
	Function   *string     `json:"function,omitempty"`
	Role       string      `json:"role"`
	Source     *string     `json:"source,omitempty"`
	ToolCallId *string     `json:"tool_call_id,omitempty"`
	ToolCalls  *[]ToolCall `json:"tool_calls,omitempty"`
}

// Output defines model for Output.
type Output struct {
	Choices *[]Choice `json:"choices,omitempty"`
	Model   *string   `json:"model,omitempty"`
	Usage   *Usage    `json:"usage,omitempty"`
}

// Project defines model for Project.
type Project struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Tools []Tool `json:"tools"`
}

// RegisterProjectRequest defines model for RegisterProjectRequest.
type RegisterProjectRequest struct {
	Name  string `json:"name"`
	Tools []Tool `json:"tools"`
}

// RegisterProjectResponse defines model for RegisterProjectResponse.
type RegisterProjectResponse struct {
	Id string `json:"id"`
}

// Review defines model for Review.
type Review struct {
	Id      string        `json:"id"`
	Request ReviewRequest `json:"request"`
}

// ReviewRequest defines model for ReviewRequest.
type ReviewRequest struct {
	AgentId      string       `json:"agent_id"`
	LastMessages []Message    `json:"last_messages"`
	TaskState    TaskState    `json:"task_state"`
	ToolChoices  []ToolChoice `json:"tool_choices"`
}

// ReviewResult defines model for ReviewResult.
type ReviewResult struct {
	Decision   Decision   `json:"decision"`
	Id         string     `json:"id"`
	Reasoning  string     `json:"reasoning"`
	ToolChoice ToolChoice `json:"tool_choice"`
}

// ReviewStatusResponse defines model for ReviewStatusResponse.
type ReviewStatusResponse struct {
	Id     string `json:"id"`
	Status Status `json:"status"`
}

// Status defines model for Status.
type Status string

// TaskState defines model for TaskState.
type TaskState struct {
	Completed  bool                    `json:"completed"`
	Messages   []Message               `json:"messages"`
	Metadata   *map[string]interface{} `json:"metadata,omitempty"`
	Output     Output                  `json:"output"`
	Store      *map[string]interface{} `json:"store,omitempty"`
	ToolChoice *ToolChoice             `json:"tool_choice,omitempty"`
	Tools      []Tool                  `json:"tools"`
}

// Tool defines model for Tool.
type Tool struct {
	Attributes  *map[string]interface{} `json:"attributes,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Name        string                  `json:"name"`
}

// ToolCall defines model for ToolCall.
type ToolCall struct {
	Arguments  map[string]interface{} `json:"arguments"`
	Function   string                 `json:"function"`
	Id         string                 `json:"id"`
	ParseError *string                `json:"parse_error,omitempty"`
	Type       string                 `json:"type"`
}

// ToolChoice defines model for ToolChoice.
type ToolChoice struct {
	Arguments Arguments `json:"arguments"`
	Function  string    `json:"function"`
	Id        string    `json:"id"`
	Type      string    `json:"type"`
}

// Usage defines model for Usage.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// GetLLMExplanationJSONRequestBody defines body for GetLLMExplanation for application/json ContentType.
type GetLLMExplanationJSONRequestBody = CodeSnippet

// RegisterProjectJSONRequestBody defines body for RegisterProject for application/json ContentType.
type RegisterProjectJSONRequestBody = RegisterProjectRequest

// SubmitReviewHumanJSONRequestBody defines body for SubmitReviewHuman for application/json ContentType.
type SubmitReviewHumanJSONRequestBody = ReviewRequest

// SubmitReviewLLMJSONRequestBody defines body for SubmitReviewLLM for application/json ContentType.
type SubmitReviewLLMJSONRequestBody = ReviewRequest

// SetLLMPromptJSONRequestBody defines body for SetLLMPrompt for application/json ContentType.
type SetLLMPromptJSONRequestBody = LLMPrompt

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get the API documentation
	// (GET /api/docs)
	GetSwaggerDocs(w http.ResponseWriter, r *http.Request)
	// Get an explanation and danger score for a code snippet
	// (POST /api/explain)
	GetLLMExplanation(w http.ResponseWriter, r *http.Request)
	// Get hub statistics
	// (GET /api/hub/stats)
	GetHubStats(w http.ResponseWriter, r *http.Request)
	// Get the OpenAPI schema
	// (GET /api/openapi.yaml)
	GetOpenAPI(w http.ResponseWriter, r *http.Request)
	// Get all projects
	// (GET /api/project)
	GetProjects(w http.ResponseWriter, r *http.Request)
	// Register a new project
	// (POST /api/project)
	RegisterProject(w http.ResponseWriter, r *http.Request)
	// Get a project by ID
	// (GET /api/project/{id})
	GetProjectById(w http.ResponseWriter, r *http.Request, id string)
	// Submit a review to a human supervisor
	// (POST /api/review/human)
	SubmitReviewHuman(w http.ResponseWriter, r *http.Request)
	// Get all LLM review results
	// (GET /api/review/llm)
	GetLLMReviews(w http.ResponseWriter, r *http.Request)
	// Submit a review to an LLM supervisor
	// (POST /api/review/llm)
	SubmitReviewLLM(w http.ResponseWriter, r *http.Request)
	// Get the current LLM review prompt
	// (GET /api/review/llm/prompt)
	GetLLMPrompt(w http.ResponseWriter, r *http.Request)
	// Submit a new prompt for LLM reviews
	// (POST /api/review/llm/prompt)
	SetLLMPrompt(w http.ResponseWriter, r *http.Request)
	// Get review status
	// (GET /api/review/status/{id})
	GetReviewStatus(w http.ResponseWriter, r *http.Request, id string)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetSwaggerDocs operation middleware
func (siw *ServerInterfaceWrapper) GetSwaggerDocs(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetSwaggerDocs(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetLLMExplanation operation middleware
func (siw *ServerInterfaceWrapper) GetLLMExplanation(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetLLMExplanation(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetHubStats operation middleware
func (siw *ServerInterfaceWrapper) GetHubStats(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetHubStats(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetOpenAPI operation middleware
func (siw *ServerInterfaceWrapper) GetOpenAPI(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetOpenAPI(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetProjects operation middleware
func (siw *ServerInterfaceWrapper) GetProjects(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetProjects(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// RegisterProject operation middleware
func (siw *ServerInterfaceWrapper) RegisterProject(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.RegisterProject(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetProjectById operation middleware
func (siw *ServerInterfaceWrapper) GetProjectById(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameterWithOptions("simple", "id", r.PathValue("id"), &id, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetProjectById(w, r, id)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// SubmitReviewHuman operation middleware
func (siw *ServerInterfaceWrapper) SubmitReviewHuman(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.SubmitReviewHuman(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetLLMReviews operation middleware
func (siw *ServerInterfaceWrapper) GetLLMReviews(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetLLMReviews(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// SubmitReviewLLM operation middleware
func (siw *ServerInterfaceWrapper) SubmitReviewLLM(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.SubmitReviewLLM(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetLLMPrompt operation middleware
func (siw *ServerInterfaceWrapper) GetLLMPrompt(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetLLMPrompt(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// SetLLMPrompt operation middleware
func (siw *ServerInterfaceWrapper) SetLLMPrompt(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.SetLLMPrompt(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetReviewStatus operation middleware
func (siw *ServerInterfaceWrapper) GetReviewStatus(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameterWithOptions("simple", "id", r.PathValue("id"), &id, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetReviewStatus(w, r, id)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{})
}

// ServeMux is an abstraction of http.ServeMux.
type ServeMux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type StdHTTPServerOptions struct {
	BaseURL          string
	BaseRouter       ServeMux
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, m ServeMux) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{
		BaseRouter: m,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, m ServeMux, baseURL string) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{
		BaseURL:    baseURL,
		BaseRouter: m,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options StdHTTPServerOptions) http.Handler {
	m := options.BaseRouter

	if m == nil {
		m = http.NewServeMux()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	m.HandleFunc("GET "+options.BaseURL+"/api/docs", wrapper.GetSwaggerDocs)
	m.HandleFunc("POST "+options.BaseURL+"/api/explain", wrapper.GetLLMExplanation)
	m.HandleFunc("GET "+options.BaseURL+"/api/hub/stats", wrapper.GetHubStats)
	m.HandleFunc("GET "+options.BaseURL+"/api/openapi.yaml", wrapper.GetOpenAPI)
	m.HandleFunc("GET "+options.BaseURL+"/api/project", wrapper.GetProjects)
	m.HandleFunc("POST "+options.BaseURL+"/api/project", wrapper.RegisterProject)
	m.HandleFunc("GET "+options.BaseURL+"/api/project/{id}", wrapper.GetProjectById)
	m.HandleFunc("POST "+options.BaseURL+"/api/review/human", wrapper.SubmitReviewHuman)
	m.HandleFunc("GET "+options.BaseURL+"/api/review/llm", wrapper.GetLLMReviews)
	m.HandleFunc("POST "+options.BaseURL+"/api/review/llm", wrapper.SubmitReviewLLM)
	m.HandleFunc("GET "+options.BaseURL+"/api/review/llm/prompt", wrapper.GetLLMPrompt)
	m.HandleFunc("POST "+options.BaseURL+"/api/review/llm/prompt", wrapper.SetLLMPrompt)
	m.HandleFunc("GET "+options.BaseURL+"/api/review/status/{id}", wrapper.GetReviewStatus)

	return m
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
