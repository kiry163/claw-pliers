package main

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show monitor status",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := callAPI("GET", "/status", nil)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage accounts",
}

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := callAPI("GET", "/accounts", nil)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var accountProvider string
var accountEmail string
var accountToken string

var accountsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an account",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload := map[string]any{
			"provider":   accountProvider,
			"email":      accountEmail,
			"auth_token": accountToken,
		}
		data, err := callAPI("POST", "/accounts", payload)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var accountsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an account",
	RunE: func(cmd *cobra.Command, args []string) error {
		escaped := url.PathEscape(accountEmail)
		data, err := callAPI("DELETE", fmt.Sprintf("/accounts/%s", escaped), nil)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var sendTo []string
var sendCc []string
var sendSubject string
var sendBody string

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload := map[string]any{
			"to":      sendTo,
			"cc":      sendCc,
			"subject": sendSubject,
			"body":    sendBody,
		}
		data, err := callAPI("POST", "/send", payload)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var testEmail string

var testConnCmd = &cobra.Command{
	Use:   "test-connection",
	Short: "Test IMAP connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload := map[string]any{"email": testEmail}
		data, err := callAPI("POST", "/test-connection", payload)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var latestEmail string
var latestCount int
var latestSince string

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Get latest emails",
	RunE: func(cmd *cobra.Command, args []string) error {
		query := url.Values{}
		if latestEmail != "" {
			query.Set("email", latestEmail)
		}
		if latestCount > 0 {
			query.Set("count", fmt.Sprintf("%d", latestCount))
		}
		if latestSince != "" {
			query.Set("since", latestSince)
		}
		path := "/latest"
		if encoded := query.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
		data, err := callAPI("GET", path, nil)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

func init() {
	accountsCmd.AddCommand(accountsListCmd)
	accountsCmd.AddCommand(accountsAddCmd)
	accountsCmd.AddCommand(accountsRemoveCmd)

	accountsAddCmd.Flags().StringVar(&accountProvider, "provider", "", "mail provider")
	accountsAddCmd.Flags().StringVar(&accountEmail, "email", "", "email address")
	accountsAddCmd.Flags().StringVar(&accountToken, "auth-token", "", "auth token")
	_ = accountsAddCmd.MarkFlagRequired("provider")
	_ = accountsAddCmd.MarkFlagRequired("email")
	_ = accountsAddCmd.MarkFlagRequired("auth-token")

	accountsRemoveCmd.Flags().StringVar(&accountEmail, "email", "", "email address")
	_ = accountsRemoveCmd.MarkFlagRequired("email")

	sendCmd.Flags().StringSliceVar(&sendTo, "to", nil, "recipient list")
	sendCmd.Flags().StringSliceVar(&sendCc, "cc", nil, "cc list")
	sendCmd.Flags().StringVar(&sendSubject, "subject", "", "email subject")
	sendCmd.Flags().StringVar(&sendBody, "body", "", "email body")
	_ = sendCmd.MarkFlagRequired("to")
	_ = sendCmd.MarkFlagRequired("subject")
	_ = sendCmd.MarkFlagRequired("body")

	testConnCmd.Flags().StringVar(&testEmail, "email", "", "email address")
	_ = testConnCmd.MarkFlagRequired("email")

	latestCmd.Flags().StringVar(&latestEmail, "email", "", "email address")
	latestCmd.Flags().IntVar(&latestCount, "count", 1, "number of emails to fetch")
	latestCmd.Flags().StringVar(&latestSince, "since", "", "time window (e.g. 1m, 30s)")
}
