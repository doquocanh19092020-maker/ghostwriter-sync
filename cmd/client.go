package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/machinebox/graphql"
)

type LogEntry struct {
	StartTime   time.Time
	EndTime     time.Time
	SourceIP    string
	DestIP      string
	Tool        string
	UserContext string
	Command     string
	Description string
	Output      string
	Comments    string
	Operator    string
	ExtraFields map[string]any
}

type GhostClient struct {
	graphql *graphql.Client
	apiKey  string
	oplogID int64
	ctx     context.Context
}

func NewGhostClient(domain string, token string, operationalId int64) (*GhostClient, error) {
	var (
		transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client    = &http.Client{Transport: transport}
		endpoint  = strings.TrimRight(domain, "/") + "/v1/graphql"
	)

	if domain == "" || token == "" {
		return nil, fmt.Errorf("all ghostwriter config values must be provided")
	}

	return &GhostClient{
		graphql: graphql.NewClient(endpoint, graphql.WithHTTPClient(client)),
		apiKey:  token,
		oplogID: operationalId,
		ctx:     context.Background(),
	}, nil
}

func (c *GhostClient) SendLog(id string, entry LogEntry) error {
	var resp map[string]any

	req := graphql.NewRequest(`
    mutation InsertLog (
        $oplog: bigint!, $startDate: timestamptz, $endDate: timestamptz,
        $sourceIp: String, $destIp: String, $tool: String, $userContext: String,
        $command: String, $description: String, $output: String, $comments: String,
        $operatorName: String, $entry_identifier: String!, $extraFields: jsonb!
    ) {
        insert_oplogEntry(objects: {
            oplog: $oplog,
            startDate: $startDate,
            endDate: $endDate,
            sourceIp: $sourceIp,
            destIp: $destIp,
            tool: $tool,
            userContext: $userContext,
            command: $command,
            description: $description,
            output: $output,
            comments: $comments,
            operatorName: $operatorName,
            entryIdentifier: $entry_identifier,
            extraFields: $extraFields
        }) {
            returning { id }
        }
    }`)

	req.Var("oplog", c.oplogID)
	req.Var("startDate", entry.StartTime.Format(time.RFC3339))
	req.Var("endDate", entry.EndTime.Format(time.RFC3339))
	req.Var("sourceIp", entry.SourceIP)
	req.Var("destIp", entry.DestIP)
	req.Var("tool", entry.Tool)
	req.Var("userContext", entry.UserContext)
	req.Var("command", entry.Command)
	req.Var("description", entry.Description)
	req.Var("output", entry.Output)
	req.Var("comments", entry.Comments)
	req.Var("operatorName", entry.Operator)
	req.Var("entry_identifier", id)
	req.Var("extraFields", entry.ExtraFields)

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return c.graphql.Run(c.ctx, req, &resp)
}

func (c *GhostClient) UpdateLogOutput(id string, additionalOutput string) error {
	var resp map[string]any

	// First, query for the entry ID using the entry_identifier
	queryReq := graphql.NewRequest(`
    query GetLogEntry($entry_identifier: String!) {
        oplogEntry(where: {entryIdentifier: {_eq: $entry_identifier}}) {
            id
            output
        }
    }`)

	queryReq.Var("entry_identifier", id)
	queryReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	var queryResp struct {
		OplogEntry []struct {
			ID     int    `json:"id"`
			Output string `json:"output"`
		} `json:"oplogEntry"`
	}

	if err := c.graphql.Run(c.ctx, queryReq, &queryResp); err != nil {
		return fmt.Errorf("failed to query log entry: %w", err)
	}

	if len(queryResp.OplogEntry) == 0 {
		return fmt.Errorf("no log entry found with identifier: %s", id)
	}

	// Update the entry with appended output
	existingOutput := queryResp.OplogEntry[0].Output
	newOutput := existingOutput + "\n" + additionalOutput
	entryID := queryResp.OplogEntry[0].ID

	updateReq := graphql.NewRequest(`
    mutation UpdateLogOutput($id: bigint!, $output: String!) {
        update_oplogEntry_by_pk(
            pk_columns: {id: $id},
            _set: {output: $output}
        ) {
            id
            output
        }
    }`)

	updateReq.Var("id", entryID)
	updateReq.Var("output", newOutput)
	updateReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	return c.graphql.Run(c.ctx, updateReq, &resp)
}
