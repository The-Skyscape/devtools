package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/The-Skyscape/devtools/pkg/ai"
)

// DeveloperAgent is an AI agent specialized for software development tasks
type DeveloperAgent struct {
	name        string
	description string
	llm         ai.LLMProvider
	tools       map[string]ai.Tool
	context     map[string]interface{}
}

// NewDeveloperAgent creates a new developer agent
func NewDeveloperAgent(llm ai.LLMProvider) *DeveloperAgent {
	return &DeveloperAgent{
		name:        "developer_agent",
		description: "An AI agent that assists with software development tasks including code analysis, repository management, and database queries",
		llm:         llm,
		tools:       make(map[string]ai.Tool),
		context:     make(map[string]interface{}),
	}
}

// Name returns the agent name
func (d *DeveloperAgent) Name() string {
	return d.name
}

// Description returns the agent's capabilities
func (d *DeveloperAgent) Description() string {
	return d.description
}

// SetTools configures the tools available to the agent
func (d *DeveloperAgent) SetTools(tools []ai.Tool) {
	d.tools = make(map[string]ai.Tool)
	for _, tool := range tools {
		d.tools[tool.Name()] = tool
	}
}

// SetContext sets the agent's working context
func (d *DeveloperAgent) SetContext(key string, value interface{}) {
	d.context[key] = value
}

// Execute runs the agent with the given task
func (d *DeveloperAgent) Execute(ctx context.Context, task string) (*ai.AgentResponse, error) {
	// Build system prompt with available tools
	systemPrompt := d.buildSystemPrompt()
	
	// Create messages for the LLM
	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: task},
	}
	
	// Process the task and determine if tools are needed
	response, toolCalls, err := d.processTask(ctx, messages)
	if err != nil {
		return nil, err
	}
	
	return &ai.AgentResponse{
		Content:   response,
		ToolCalls: toolCalls,
		Metadata: map[string]interface{}{
			"agent":   d.name,
			"context": d.context,
		},
	}, nil
}

// Chat handles a conversational interaction
func (d *DeveloperAgent) Chat(ctx context.Context, messages []ai.Message) (*ai.AgentResponse, error) {
	// Add system prompt if not present
	if len(messages) == 0 || messages[0].Role != "system" {
		systemPrompt := d.buildSystemPrompt()
		messages = append([]ai.Message{{Role: "system", Content: systemPrompt}}, messages...)
	}
	
	// Process the conversation
	response, toolCalls, err := d.processTask(ctx, messages)
	if err != nil {
		return nil, err
	}
	
	return &ai.AgentResponse{
		Content:   response,
		ToolCalls: toolCalls,
		Metadata: map[string]interface{}{
			"agent":   d.name,
			"context": d.context,
		},
	}, nil
}

// processTask processes a task and executes tools if needed
func (d *DeveloperAgent) processTask(ctx context.Context, messages []ai.Message) (string, []ai.ToolCall, error) {
	// First, get the LLM's response
	req := ai.CompletionRequest{
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2048,
	}
	
	resp, err := d.llm.Complete(ctx, req)
	if err != nil {
		return "", nil, err
	}
	
	// Parse the response to check if tools are needed
	toolCalls := d.parseToolCalls(resp.Content)
	
	// Execute any tool calls
	for i, toolCall := range toolCalls {
		tool, exists := d.tools[toolCall.Tool]
		if !exists {
			toolCalls[i].Result = fmt.Sprintf("Error: Tool '%s' not found", toolCall.Tool)
			continue
		}
		
		result, err := tool.Execute(ctx, toolCall.Parameters)
		if err != nil {
			toolCalls[i].Result = fmt.Sprintf("Error executing tool: %v", err)
		} else {
			toolCalls[i].Result = result
		}
	}
	
	// If tools were executed, generate a final response incorporating the results
	if len(toolCalls) > 0 {
		// Add tool results to the conversation
		toolResultsMsg := d.formatToolResults(toolCalls)
		messages = append(messages, ai.Message{
			Role:    "assistant",
			Content: toolResultsMsg,
		})
		
		// Get final response
		finalReq := ai.CompletionRequest{
			Messages:    messages,
			Temperature: 0.7,
			MaxTokens:   2048,
		}
		
		finalResp, err := d.llm.Complete(ctx, finalReq)
		if err != nil {
			return resp.Content, toolCalls, nil // Return original response if final fails
		}
		
		return finalResp.Content, toolCalls, nil
	}
	
	return resp.Content, toolCalls, nil
}

// buildSystemPrompt creates the system prompt with tool descriptions
func (d *DeveloperAgent) buildSystemPrompt() string {
	prompt := `You are a developer assistant AI agent with access to various tools for software development tasks.

Your capabilities include:
- Analyzing code and providing insights
- Managing repositories and issues
- Querying databases
- Searching and understanding codebases
- Generating documentation and tests

Available tools:`

	for name, tool := range d.tools {
		prompt += fmt.Sprintf("\n\n- **%s**: %s", name, tool.Description())
		
		// Add parameter details
		if params := tool.Parameters(); params != nil {
			if props, ok := params["properties"].(map[string]interface{}); ok {
				prompt += "\n  Parameters:"
				for param, details := range props {
					if detailMap, ok := details.(map[string]interface{}); ok {
						if desc, ok := detailMap["description"].(string); ok {
							prompt += fmt.Sprintf("\n    - %s: %s", param, desc)
						}
					}
				}
			}
		}
	}
	
	// Add context information if available
	if repoID, ok := d.context["repo_id"].(string); ok {
		prompt += fmt.Sprintf("\n\nCurrent repository context: %s", repoID)
	}
	
	prompt += `

When you need to use a tool, respond with a tool call in this format:
TOOL_CALL: tool_name
PARAMETERS:
{
  "param1": "value1",
  "param2": "value2"
}
END_TOOL_CALL

You can make multiple tool calls in a single response. After executing tools, provide a comprehensive answer based on the results.`

	return prompt
}

// parseToolCalls extracts tool calls from the LLM response
func (d *DeveloperAgent) parseToolCalls(content string) []ai.ToolCall {
	var toolCalls []ai.ToolCall
	
	// Look for tool call patterns in the response
	lines := strings.Split(content, "\n")
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		
		if strings.HasPrefix(line, "TOOL_CALL:") {
			toolName := strings.TrimSpace(strings.TrimPrefix(line, "TOOL_CALL:"))
			
			// Look for parameters
			var params map[string]interface{}
			i++
			if i < len(lines) && strings.TrimSpace(lines[i]) == "PARAMETERS:" {
				i++
				// Collect JSON until END_TOOL_CALL
				var jsonLines []string
				for i < len(lines) && strings.TrimSpace(lines[i]) != "END_TOOL_CALL" {
					jsonLines = append(jsonLines, lines[i])
					i++
				}
				
				// Parse JSON parameters
				jsonStr := strings.Join(jsonLines, "\n")
				json.Unmarshal([]byte(jsonStr), &params)
			}
			
			toolCalls = append(toolCalls, ai.ToolCall{
				Tool:       toolName,
				Parameters: params,
			})
		}
		i++
	}
	
	return toolCalls
}

// formatToolResults formats tool execution results for the LLM
func (d *DeveloperAgent) formatToolResults(toolCalls []ai.ToolCall) string {
	var results []string
	
	for _, call := range toolCalls {
		resultStr := fmt.Sprintf("Tool: %s\n", call.Tool)
		
		// Format the result based on its type
		switch r := call.Result.(type) {
		case string:
			resultStr += fmt.Sprintf("Result: %s", r)
		case map[string]interface{}:
			jsonBytes, _ := json.MarshalIndent(r, "", "  ")
			resultStr += fmt.Sprintf("Result:\n%s", string(jsonBytes))
		default:
			resultStr += fmt.Sprintf("Result: %v", r)
		}
		
		results = append(results, resultStr)
	}
	
	return "Tool execution results:\n\n" + strings.Join(results, "\n\n---\n\n")
}

// CodeReviewAgent is specialized for code review tasks
type CodeReviewAgent struct {
	*DeveloperAgent
}

// NewCodeReviewAgent creates a new code review agent
func NewCodeReviewAgent(llm ai.LLMProvider) *CodeReviewAgent {
	agent := &CodeReviewAgent{
		DeveloperAgent: NewDeveloperAgent(llm),
	}
	agent.name = "code_review_agent"
	agent.description = "An AI agent specialized in code review, finding bugs, suggesting improvements, and ensuring code quality"
	return agent
}

// ReviewPullRequest reviews a pull request
func (c *CodeReviewAgent) ReviewPullRequest(ctx context.Context, prDiff string) (*ai.AgentResponse, error) {
	task := fmt.Sprintf(`Please review this pull request diff and provide:
1. Summary of changes
2. Potential issues or bugs
3. Code quality observations
4. Security concerns if any
5. Suggestions for improvement

Diff:
%s`, prDiff)

	return c.Execute(ctx, task)
}

// RepositoryAssistant helps with repository management
type RepositoryAssistant struct {
	*DeveloperAgent
}

// NewRepositoryAssistant creates a new repository assistant
func NewRepositoryAssistant(llm ai.LLMProvider) *RepositoryAssistant {
	agent := &RepositoryAssistant{
		DeveloperAgent: NewDeveloperAgent(llm),
	}
	agent.name = "repository_assistant"
	agent.description = "An AI assistant for repository management, helping with issues, documentation, and codebase understanding"
	return agent
}

// AnalyzeRepository provides a comprehensive repository analysis
func (r *RepositoryAssistant) AnalyzeRepository(ctx context.Context, repoID string) (*ai.AgentResponse, error) {
	r.SetContext("repo_id", repoID)
	
	task := fmt.Sprintf(`Please provide a comprehensive analysis of repository %s:
1. Use the repository_tool to explore the file structure
2. Identify the main technologies and frameworks used
3. Analyze the code organization and architecture
4. Check for any potential issues or improvements
5. Provide a summary of the repository's purpose and functionality`, repoID)

	return r.Execute(ctx, task)
}