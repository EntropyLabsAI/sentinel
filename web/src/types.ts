/**
 * Generated by orval v7.1.1 🍺
 * Do not edit manually.
 * Sentinel API
 * OpenAPI spec version: 1.0.0
 */
import axios from 'axios'
import type {
  AxiosRequestConfig,
  AxiosResponse
} from 'axios'
export type GetReviewResultParams = {
id: string;
};

export type HubStatsReviewDistribution = {[key: string]: number};

export type HubStatsAssignedReviews = {[key: string]: number};

export interface HubStats {
  assigned_reviews: HubStatsAssignedReviews;
  busy_clients: number;
  completed_reviews: number;
  connected_clients: number;
  free_clients: number;
  queued_reviews: number;
  review_distribution: HubStatsReviewDistribution;
  stored_reviews: number;
}

export type Status = typeof Status[keyof typeof Status];


// eslint-disable-next-line @typescript-eslint/no-redeclare
export const Status = {
  queued: 'queued',
  processing: 'processing',
  completed: 'completed',
  timeout: 'timeout',
} as const;

export type Decision = typeof Decision[keyof typeof Decision];


// eslint-disable-next-line @typescript-eslint/no-redeclare
export const Decision = {
  approve: 'approve',
  reject: 'reject',
  escalate: 'escalate',
  terminate: 'terminate',
  modify: 'modify',
} as const;

export interface Usage {
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
}

export type ToolCallArguments = { [key: string]: unknown };

export interface ToolCall {
  arguments: ToolCallArguments;
  function: string;
  id: string;
  parse_error?: string;
  type: string;
}

export interface AssistantMessage {
  content: string;
  role: string;
  source?: string;
  tool_calls?: ToolCall[];
}

export interface Choice {
  message: AssistantMessage;
  stop_reason?: string;
}

export interface Output {
  choices?: Choice[];
  model?: string;
  usage?: Usage;
}

export interface Arguments {
  cmd?: string;
  code?: string;
}

export interface ToolChoice {
  arguments: Arguments;
  function: string;
  id: string;
  type: string;
}

export type ToolAttributes = { [key: string]: unknown };

export interface Tool {
  attributes?: ToolAttributes;
  description?: string;
  name: string;
}

export interface Message {
  content: string;
  function?: string;
  role: string;
  source?: string;
  tool_call_id?: string;
  tool_calls?: ToolCall[];
}

export type TaskStateStore = { [key: string]: unknown };

export type TaskStateMetadata = { [key: string]: unknown };

export interface TaskState {
  completed: boolean;
  messages: Message[];
  metadata?: TaskStateMetadata;
  output: Output;
  store?: TaskStateStore;
  tool_choice?: ToolChoice;
  tools: Tool[];
}

export interface ErrorResponse {
  status: string;
}

export interface ReviewResult {
  decision: Decision;
  id: string;
  reasoning: string;
  tool_choice: ToolChoice;
}

export interface ReviewStatusResponse {
  id: string;
  status: Status;
}

export interface ReviewRequest {
  agent_id: string;
  last_messages: Message[];
  task_state: TaskState;
  tool_choices: ToolChoice[];
}

export interface Review {
  id: string;
  request: ReviewRequest;
}

export interface LLMExplanation {
  explanation: string;
  score: number;
}

export interface CodeSnippet {
  text: string;
}





  /**
 * @summary Submit a review request for human review
 */
export const submitReview = <TData = AxiosResponse<ReviewStatusResponse>>(
    reviewRequest: ReviewRequest, options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.post(
      `/api/review/human`,
      reviewRequest,options
    );
  }

/**
 * @summary Submit a review request for LLM review
 */
export const submitReviewLLM = <TData = AxiosResponse<ReviewStatusResponse>>(
    reviewRequest: ReviewRequest, options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.post(
      `/api/review/llm`,
      reviewRequest,options
    );
  }

/**
 * @summary Get the status of a review request
 */
export const getReviewResult = <TData = AxiosResponse<ReviewResult>>(
    params: GetReviewResultParams, options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.get(
      `/api/review/status`,{
    ...options,
        params: {...params, ...options?.params},}
    );
  }

/**
 * @summary Get all LLM review results
 */
export const getLLMReviews = <TData = AxiosResponse<Review[]>>(
     options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.get(
      `/api/review/llm/list`,options
    );
  }

/**
 * @summary Get hub statistics
 */
export const getHubStats = <TData = AxiosResponse<HubStats>>(
     options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.get(
      `/api/stats`,options
    );
  }

/**
 * @summary Get an explanation and danger score for a code snippet
 */
export const getLLMExplanation = <TData = AxiosResponse<LLMExplanation>>(
    codeSnippet: CodeSnippet, options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.post(
      `/api/explain`,
      codeSnippet,options
    );
  }

export type SubmitReviewResult = AxiosResponse<ReviewStatusResponse>
export type SubmitReviewLLMResult = AxiosResponse<ReviewStatusResponse>
export type GetReviewResultResult = AxiosResponse<ReviewResult>
export type GetLLMReviewsResult = AxiosResponse<Review[]>
export type GetHubStatsResult = AxiosResponse<HubStats>
export type GetLLMExplanationResult = AxiosResponse<LLMExplanation>
