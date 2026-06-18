package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

const (
	ExitSuccess = 0
	ExitUsage   = 1
	ExitAPI     = 2
	ExitError   = 3
	ExitConfig  = 3
)

// Root is the top-level command dispatcher for cursor-agent-cli.
type Root struct {
	stdout        io.Writer
	stderr        io.Writer
	clientFactory func() (cursor.Client, error)
}

// NewRoot creates a Root command with stdout and stderr wired to os.Stdout and os.Stderr.
func NewRoot() *Root {
	return &Root{
		stdout:        os.Stdout,
		stderr:        os.Stderr,
		clientFactory: cursor.ClientFromEnv,
	}
}

func (r *Root) apiClient() (cursor.Client, error) {
	if r.clientFactory != nil {
		return r.clientFactory()
	}
	return cursor.ClientFromEnv()
}

// Run dispatches to a subcommand or prints the default hello-world response.
func (r *Root) Run(args []string) int {
	if len(args) == 0 {
		return r.runHello()
	}

	switch args[0] {
	case "help", "-h", "--help":
		return r.runHelp(args[1:])
	case "models":
		return r.runModels(args[1:])
	case "list":
		return r.runList(args[1:])
	case "create":
		return r.runCreate(args[1:])
	case "run":
		return r.runRun(args[1:])
	case "status":
		return r.runStatus(args[1:])
	default:
		return r.runUnknown(args[0])
	}
}

func (r *Root) runModels(args []string) int {
	fs := flag.NewFlagSet("models", flag.ContinueOnError)
	fs.SetOutput(r.stderr)
	fs.Usage = modelsUsage(r.stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitSuccess
		}
		return r.fail(ExitUsage, err)
	}

	client, err := r.apiClient()
	if err != nil {
		return r.fail(ExitConfig, err)
	}

	resp, err := listModels(context.Background(), client)
	if err != nil {
		return r.fail(ExitAPI, err)
	}
	return r.writeJSON(resp)
}

func (r *Root) runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(r.stderr)
	limit := fs.Int("limit", 20, "maximum number of agents to return")
	fs.Usage = listUsage(r.stderr, fs)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitSuccess
		}
		return r.fail(ExitUsage, err)
	}
	if *limit <= 0 {
		return r.fail(ExitUsage, fmt.Errorf("--limit must be greater than 0, got %d", *limit))
	}

	client, err := r.apiClient()
	if err != nil {
		return r.fail(ExitConfig, err)
	}

	resp, err := listAgents(context.Background(), client, *limit)
	if err != nil {
		return r.fail(ExitAPI, err)
	}
	return r.writeJSON(resp)
}

func (r *Root) runCreate(args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(r.stderr)
	repo := fs.String("repo", "", "GitHub repository URL (required)")
	prompt := fs.String("prompt", "", "task prompt for the agent (required)")
	branch := fs.String("branch", "main", "branch name or commit SHA to use as the starting point")
	fs.Usage = func() {
		fmt.Fprintln(r.stderr, "Usage: cursor-agent-cli create [flags]")
		fmt.Fprintln(r.stderr)
		fmt.Fprintln(r.stderr, "Flags:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitSuccess
		}
		return r.fail(ExitUsage, err)
	}
	if strings.TrimSpace(*repo) == "" {
		return r.fail(ExitUsage, fmt.Errorf("--repo is required"))
	}
	if strings.TrimSpace(*prompt) == "" {
		return r.fail(ExitUsage, fmt.Errorf("--prompt is required"))
	}

	branchRef := strings.TrimSpace(*branch)
	autoCreatePR := true
	req := cursor.CreateAgentRequest{
		Prompt: cursor.AgentPrompt{Text: *prompt},
		Repos: []cursor.AgentRepo{
			{
				URL:         strings.TrimSpace(*repo),
				StartingRef: &branchRef,
			},
		},
		AutoCreatePR: &autoCreatePR,
	}

	client, err := r.apiClient()
	if err != nil {
		return r.fail(ExitConfig, err)
	}

	resp, err := createAgent(context.Background(), client, req)
	if err != nil {
		return r.fail(ExitAPI, err)
	}
	return r.writeJSON(resp)
}

func (r *Root) runRun(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(r.stderr)
	prompt := fs.String("prompt", "", "modification instruction text (required)")
	fs.Usage = func() {
		fmt.Fprintln(r.stderr, "Usage: cursor-agent-cli run <agent_id> [flags]")
		fmt.Fprintln(r.stderr)
		fmt.Fprintln(r.stderr, "Flags:")
		fs.PrintDefaults()
	}

	var agentID string
	flagArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		agentID = strings.TrimSpace(args[0])
		flagArgs = args[1:]
	}

	if err := fs.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitSuccess
		}
		return r.fail(ExitUsage, err)
	}

	if agentID == "" {
		agentID = strings.TrimSpace(fs.Arg(0))
	}
	if agentID == "" {
		return r.fail(ExitUsage, fmt.Errorf("agent_id is required"))
	}
	if strings.TrimSpace(*prompt) == "" {
		return r.fail(ExitUsage, fmt.Errorf("--prompt is required"))
	}

	req := cursor.CreateRunRequest{
		Prompt: cursor.AgentPrompt{Text: *prompt},
	}

	client, err := r.apiClient()
	if err != nil {
		return r.fail(ExitConfig, err)
	}

	resp, err := createRun(context.Background(), client, agentID, req)
	if err != nil {
		return r.fail(ExitAPI, err)
	}
	return r.writeJSON(resp)
}

func (r *Root) runStatus(args []string) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(r.stderr)
	watch := fs.Bool("watch", false, "poll until the run reaches a terminal status")
	interval := fs.Int("interval", 15, "polling interval in seconds")
	timeout := fs.Int("timeout", 0, "maximum wait time in seconds (0 = no limit)")
	fs.Usage = func() {
		fmt.Fprintln(r.stderr, "Usage: cursor-agent-cli status <agent_id> <run_id> [flags]")
		fmt.Fprintln(r.stderr)
		fmt.Fprintln(r.stderr, "Flags:")
		fs.PrintDefaults()
	}

	var agentID, runID string
	agentIDFromArgs := false
	runIDFromArgs := false
	flagArgs := args
	if len(flagArgs) > 0 && !strings.HasPrefix(flagArgs[0], "-") {
		agentID = strings.TrimSpace(flagArgs[0])
		agentIDFromArgs = true
		flagArgs = flagArgs[1:]
	}
	if len(flagArgs) > 0 && !strings.HasPrefix(flagArgs[0], "-") {
		runID = strings.TrimSpace(flagArgs[0])
		runIDFromArgs = true
		flagArgs = flagArgs[1:]
	}

	if err := fs.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitSuccess
		}
		return r.writeStatusResponse(newCLIOnlyStatus(cliStateUsageError, ExitUsage, err))
	}

	if agentID == "" && fs.NArg() >= 1 {
		agentID = strings.TrimSpace(fs.Arg(0))
	}
	if runID == "" {
		if fs.NArg() >= 2 {
			runID = strings.TrimSpace(fs.Arg(1))
		} else if fs.NArg() == 1 && agentIDFromArgs && !runIDFromArgs {
			runID = strings.TrimSpace(fs.Arg(0))
		}
	}
	if agentID == "" {
		return r.writeStatusResponse(newCLIOnlyStatus(cliStateUsageError, ExitUsage, fmt.Errorf("agent_id is required")))
	}
	if runID == "" {
		return r.writeStatusResponse(newCLIOnlyStatus(cliStateUsageError, ExitUsage, fmt.Errorf("run_id is required")))
	}
	if *interval < 5 {
		return r.writeStatusResponse(newCLIOnlyStatus(cliStateUsageError, ExitUsage, fmt.Errorf("--interval must be at least 5 seconds, got %d", *interval)))
	}
	if *timeout < 0 {
		return r.writeStatusResponse(newCLIOnlyStatus(cliStateUsageError, ExitUsage, fmt.Errorf("--timeout must be greater than or equal to 0, got %d", *timeout)))
	}

	client, err := r.apiClient()
	if err != nil {
		return r.writeStatusResponse(newCLIOnlyStatus(cliStateConfigError, ExitConfig, err))
	}

	ctx := context.Background()
	intervalDur := time.Duration(*interval) * time.Second
	timeoutDur := time.Duration(*timeout) * time.Second

	var outcome statusOutcome
	if *watch {
		outcome = waitForRunStatus(ctx, client, agentID, runID, intervalDur, timeoutDur)
	} else {
		outcome = getRunStatus(ctx, client, agentID, runID)
	}
	return r.writeStatusResponse(statusResponseFromOutcome(agentID, runID, outcome))
}

func (r *Root) fail(code int, err error) int {
	fmt.Fprintf(r.stderr, "error: %v\n", err)
	return code
}

func (r *Root) runHello() int {
	return r.writeJSON(map[string]string{
		"message": "hello from cursor-agent-cli",
	})
}

func (r *Root) runHelp(_ []string) int {
	fs := flag.NewFlagSet("cursor-agent-cli", flag.ContinueOnError)
	fs.SetOutput(r.stderr)
	fs.Usage = func() {
		fmt.Fprintln(r.stderr, "Usage: cursor-agent-cli [command]")
		fmt.Fprintln(r.stderr)
		fmt.Fprintln(r.stderr, "Commands:")
		fmt.Fprintln(r.stderr, "  help     Show usage information")
		fmt.Fprintln(r.stderr, "  models   List available models")
		fmt.Fprintln(r.stderr, "  list     List agents")
		fmt.Fprintln(r.stderr, "  create   Create a Cloud Agent")
		fmt.Fprintln(r.stderr, "  run      Start a new run on an existing agent")
		fmt.Fprintln(r.stderr, "  status   Get the status of an agent run")
	}
	fs.Usage()
	return ExitSuccess
}

func (r *Root) runUnknown(name string) int {
	fmt.Fprintf(r.stderr, "unknown command: %s\n", name)
	return ExitUsage
}

func (r *Root) writeJSON(v any) int {
	enc := json.NewEncoder(r.stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(r.stderr, "error: %v\n", err)
		return ExitError
	}
	return ExitSuccess
}

func (r *Root) writeStatusResponse(resp StatusResponse) int {
	code := resp.CLI.ExitCode
	if resp.CLI.Error != nil {
		fmt.Fprintf(r.stderr, "error: %s\n", *resp.CLI.Error)
	}
	enc := json.NewEncoder(r.stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		fmt.Fprintf(r.stderr, "error: %v\n", err)
		return ExitError
	}
	return code
}
