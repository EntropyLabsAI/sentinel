package main

// Constants for message roles
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// ReviewRequest represents the review data structure
type ReviewRequest struct {
	ID             string        `json:"id"`
	Context        ReviewContext `json:"context"`
	ProposedAction string        `json:"proposed_action"`
}

// ReviewerResponse represents the response from the reviewer
type ReviewerResponse struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
}

// ReviewContext represents the entire context of a review.
type ReviewContext struct {
	Messages   []Message              `json:"messages"`
	Tools      []Tool                 `json:"tools"`
	ToolChoice *ToolChoice            `json:"tool_choice,omitempty"`
	Store      map[string]interface{} `json:"store"`
	Output     Output                 `json:"output"`
	Completed  bool                   `json:"completed"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Message represents each message in the context.
type Message struct {
	Content    string     `json:"content"`
	Role       string     `json:"role"`
	Source     string     `json:"source,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // New Field
	Function   string     `json:"function,omitempty"`     // New Field
}

// Tool represents a tool available in the context.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// ToolChoice represents the tool selection details.
type ToolChoice struct {
	ID        string    `json:"id"`
	Function  string    `json:"function"`
	Arguments Arguments `json:"arguments"`
	Type      string    `json:"type"`
}

// Arguments represents the arguments passed to a tool function.
type Arguments struct {
	Cmd string `json:"cmd"`
	// Add other fields if there are more arguments
}

// Output represents the output section of the context.
type Output struct {
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents each choice in the output.
type Choice struct {
	Message    AssistantMessage `json:"message"`
	StopReason string           `json:"stop_reason"`
}

// AssistantMessage represents the assistant's message within a choice.
type AssistantMessage struct {
	Content   string     `json:"content"`
	Source    string     `json:"source"`
	Role      string     `json:"role"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a function call made by a tool.
type ToolCall struct {
	ID        string                 `json:"id"`
	Function  string                 `json:"function"`
	Arguments map[string]interface{} `json:"arguments"`
	Type      string                 `json:"type"`
}

// Usage represents the token usage statistics.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
