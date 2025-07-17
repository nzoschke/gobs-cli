package main

import (
	"fmt"

	"github.com/andreykaipov/goobs/api/requests/config"
)

// SettingsCmd handles settings management.
type SettingsCmd struct {
	Show ShowCmd `help:"Show video settings." cmd:""`
}

// ShowCmd shows the video settings.
type ShowCmd struct{}

// Run executes the show command.
func (cmd *ShowCmd) Run(ctx *context) error {
	// Get video settings
	videoResp, err := ctx.Client.Config.GetVideoSettings()
	if err != nil {
		return fmt.Errorf("failed to get video settings: %w", err)
	}

	fmt.Fprintf(ctx.Out, "Video Settings:\n")
	fmt.Fprintf(ctx.Out, "  Base Width: %.0f\n", videoResp.BaseWidth)
	fmt.Fprintf(ctx.Out, "  Base Height: %.0f\n", videoResp.BaseHeight)
	fmt.Fprintf(ctx.Out, "  Output Width: %.0f\n", videoResp.OutputWidth)
	fmt.Fprintf(ctx.Out, "  Output Height: %.0f\n", videoResp.OutputHeight)
	fmt.Fprintf(ctx.Out, "  FPS Numerator: %.0f\n", videoResp.FpsNumerator)
	fmt.Fprintf(ctx.Out, "  FPS Denominator: %.0f\n", videoResp.FpsDenominator)

	fmt.Fprintf(ctx.Out, "\nRecord Settings:\n")
	dirResp, dirErr := ctx.Client.Config.GetRecordDirectory()
	if dirErr == nil {
		fmt.Fprintf(ctx.Out, "  Record Directory: %s\n", dirResp.RecordDirectory)
	}

	// Get profile parameters
	fmt.Fprintf(ctx.Out, "\nProfile Parameters:\n")

	// Common profile parameters to display
	params := []struct {
		category string
		name     string
		label    string
	}{
		{"Output", "Mode", "Output Mode"},
		{"SimpleOutput", "RecFormat", "Recording Format"},
		{"SimpleOutput", "RecEncoder", "Recording Encoder"},
		{"SimpleOutput", "RecQuality", "Recording Quality"},
		{"SimpleOutput", "StreamEncoder", "Streaming Encoder"},
		{"AdvOut", "RecType", "Advanced Recording Type"},
		{"AdvOut", "FFFormat", "Advanced Recording Format"},
		{"AdvOut", "RecFormat2", "Advanced Recording Format"},
		{"AdvOut", "RecAudioEncoder", "Advanced Audio Format"},
	}

	for _, param := range params {
		resp, err := ctx.Client.Config.GetProfileParameter(
			config.NewGetProfileParameterParams().
				WithParameterCategory(param.category).
				WithParameterName(param.name),
		)
		if err == nil && resp.ParameterValue != "" {
			fmt.Fprintf(ctx.Out, "  %s: %s\n", param.label, resp.ParameterValue)
		}
	}

	return nil
}
