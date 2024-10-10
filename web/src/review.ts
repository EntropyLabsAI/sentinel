export interface HubStats {
  connected_clients: number;
  queued_reviews: number;
  stored_reviews: number;
  free_clients: number;
  busy_clients: number;
  assigned_reviews: { [key: string]: number };
  review_distribution: { [key: number]: number };
  completed_reviews: number;
}

export interface ReviewRequest {
  agent_id: string;
  request_id: string;
  task_state: TaskState;
  tool_choices: ToolChoice[];
  last_messages: Message[];
}

export interface ReviewResponse {
  id: string;
  decision: string;
  tool_choice: ToolChoice;
}

export interface TaskState {
  messages: Message[];
  tools: Tool[];
  tool_choice?: ToolChoice;
  store: Record<string, any>;
  output: Output;
  completed: boolean;
  metadata: Record<string, any>;
}

export interface Message {
  content: string;
  role: string;
  source?: string;
  tool_calls?: ToolCall[];
  tool_call_id?: string;
  function?: string;
}

export interface Tool {
  name: string;
  description?: string;
  attributes?: Record<string, any>;
}

export interface ToolChoice {
  id: string;
  function: string;
  arguments: Arguments;
  type: string;
}

export interface Arguments {
  cmd: string;
  code: string;
  // Add other fields if there are more arguments
}

export interface Output {
  model: string;
  choices: Choice[];
  usage: Usage;
}

export interface Choice {
  message: AssistantMessage;
  stop_reason: string;
}

export interface AssistantMessage {
  content: string;
  source: string;
  role: string;
  tool_calls?: ToolCall[];
}

export interface ToolCall {
  id: string;
  function: string;
  arguments: Record<string, any>;
  type: string;
}

export interface Usage {
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
}

// Constants for message roles
export const RoleSystem = "system";
export const RoleUser = "user";
export const RoleAssistant = "assistant";
export const RoleTool = "tool";
