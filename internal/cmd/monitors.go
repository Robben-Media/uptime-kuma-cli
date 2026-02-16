package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/uptime-kuma-cli/internal/outfmt"
	"github.com/builtbyrobben/uptime-kuma-cli/internal/uptimekuma"
)

type MonitorsCmd struct {
	List       MonitorsListCmd       `cmd:"" help:"List all monitors"`
	Get        MonitorsGetCmd        `cmd:"" help:"Get a monitor by ID"`
	Create     MonitorsCreateCmd     `cmd:"" help:"Create a new monitor"`
	Heartbeats MonitorsHeartbeatsCmd `cmd:"" help:"Get heartbeats for a monitor"`
	Pause      MonitorsPauseCmd      `cmd:"" help:"Pause a monitor"`
	Resume     MonitorsResumeCmd     `cmd:"" help:"Resume a paused monitor"`
	Delete     MonitorsDeleteCmd     `cmd:"" help:"Delete a monitor"`
}

type MonitorsListCmd struct{}

func (cmd *MonitorsListCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	monitors, err := client.ListMonitors(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, monitors)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE", "STATUS", "URL"}
		rows := make([][]string, len(monitors))

		for i, m := range monitors {
			u := m.URL
			if u == "" {
				u = m.Hostname
			}

			rows[i] = []string{
				strconv.Itoa(m.ID),
				m.Name,
				m.Type,
				statusString(m.Status),
				u,
			}
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	for _, m := range monitors {
		u := m.URL
		if u == "" {
			u = m.Hostname
		}

		fmt.Fprintf(os.Stdout, "[%d] %s (%s) - %s  %s\n", m.ID, m.Name, m.Type, statusString(m.Status), u)
	}

	return nil
}

type MonitorsGetCmd struct {
	ID int `arg:"" help:"Monitor ID"`
}

func (cmd *MonitorsGetCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	monitor, err := client.GetMonitor(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, monitor)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE", "STATUS", "URL", "INTERVAL"}
		u := monitor.URL
		if u == "" {
			u = monitor.Hostname
		}

		rows := [][]string{{
			strconv.Itoa(monitor.ID),
			monitor.Name,
			monitor.Type,
			statusString(monitor.Status),
			u,
			strconv.Itoa(monitor.Interval),
		}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	u := monitor.URL
	if u == "" {
		u = monitor.Hostname
	}

	fmt.Fprintf(os.Stdout, "ID:       %d\n", monitor.ID)
	fmt.Fprintf(os.Stdout, "Name:     %s\n", monitor.Name)
	fmt.Fprintf(os.Stdout, "Type:     %s\n", monitor.Type)
	fmt.Fprintf(os.Stdout, "Status:   %s\n", statusString(monitor.Status))
	fmt.Fprintf(os.Stdout, "URL:      %s\n", u)
	fmt.Fprintf(os.Stdout, "Interval: %ds\n", monitor.Interval)
	fmt.Fprintf(os.Stdout, "Active:   %t\n", monitor.Active)

	if monitor.Description != "" {
		fmt.Fprintf(os.Stdout, "Desc:     %s\n", monitor.Description)
	}

	return nil
}

type MonitorsCreateCmd struct {
	Name     string `help:"Monitor name" required:""`
	Type     string `help:"Monitor type (http, port, ping, keyword, dns, docker, push, steam, mqtt, sqlserver, postgres, mysql, mongodb, radius)" required:""`
	URL      string `help:"URL to monitor (for http/keyword types)"`
	Hostname string `help:"Hostname to monitor (for ping/port/dns types)"`
	Interval int    `help:"Check interval in seconds" default:"60"`
}

func (cmd *MonitorsCreateCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	input := uptimekuma.CreateMonitorInput{
		Name:     cmd.Name,
		Type:     cmd.Type,
		URL:      cmd.URL,
		Hostname: cmd.Hostname,
		Interval: cmd.Interval,
	}

	monitor, err := client.CreateMonitor(ctx, input)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, monitor)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE"}
		rows := [][]string{{
			strconv.Itoa(monitor.ID),
			monitor.Name,
			monitor.Type,
		}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stdout, "Created monitor [%d] %s (%s)\n", monitor.ID, monitor.Name, monitor.Type)

	return nil
}

type MonitorsHeartbeatsCmd struct {
	ID int `arg:"" help:"Monitor ID"`
}

func (cmd *MonitorsHeartbeatsCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	beats, err := client.GetHeartbeats(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, beats)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "STATUS", "TIME", "PING", "MSG"}
		rows := make([][]string, len(beats))

		for i, b := range beats {
			rows[i] = []string{
				strconv.Itoa(b.ID),
				statusString(b.Status),
				b.Time,
				strconv.Itoa(b.Ping),
				b.Msg,
			}
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	for _, b := range beats {
		fmt.Fprintf(os.Stdout, "[%d] %s - %s (ping: %dms) %s\n", b.ID, statusString(b.Status), b.Time, b.Ping, b.Msg)
	}

	return nil
}

type MonitorsPauseCmd struct {
	ID int `arg:"" help:"Monitor ID"`
}

func (cmd *MonitorsPauseCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	if err := client.PauseMonitor(ctx, cmd.ID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"status":     "success",
			"monitor_id": cmd.ID,
			"message":    "Monitor paused",
		})
	}

	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "MONITOR_ID"}, [][]string{{"paused", strconv.Itoa(cmd.ID)}})
	}

	fmt.Fprintf(os.Stdout, "Monitor %d paused\n", cmd.ID)

	return nil
}

type MonitorsResumeCmd struct {
	ID int `arg:"" help:"Monitor ID"`
}

func (cmd *MonitorsResumeCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	if err := client.ResumeMonitor(ctx, cmd.ID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"status":     "success",
			"monitor_id": cmd.ID,
			"message":    "Monitor resumed",
		})
	}

	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "MONITOR_ID"}, [][]string{{"resumed", strconv.Itoa(cmd.ID)}})
	}

	fmt.Fprintf(os.Stdout, "Monitor %d resumed\n", cmd.ID)

	return nil
}

type MonitorsDeleteCmd struct {
	ID int `arg:"" help:"Monitor ID"`
}

func (cmd *MonitorsDeleteCmd) Run(ctx context.Context) error {
	client, err := getUptimeKumaClient()
	if err != nil {
		return err
	}

	if err := client.DeleteMonitor(ctx, cmd.ID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"status":     "success",
			"monitor_id": cmd.ID,
			"message":    "Monitor deleted",
		})
	}

	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "MONITOR_ID"}, [][]string{{"deleted", strconv.Itoa(cmd.ID)}})
	}

	fmt.Fprintf(os.Stdout, "Monitor %d deleted\n", cmd.ID)

	return nil
}
