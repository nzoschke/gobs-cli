// nolint: misspell
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// InputCmd provides commands to manage inputs in OBS Studio.
type InputCmd struct {
	List   InputListCmd   `cmd:"" help:"List all inputs." aliases:"ls"`
	Mute   InputMuteCmd   `cmd:"" help:"Mute input."      aliases:"m"`
	Unmute InputUnmuteCmd `cmd:"" help:"Unmute input."    aliases:"um"`
	Toggle InputToggleCmd `cmd:"" help:"Toggle input."    aliases:"tg"`
	Create InputCreateCmd `cmd:"" help:"Create input."    aliases:"c"`
	Show   InputShowCmd   `cmd:"" help:"Show input details." aliases:"s"`
}

// InputListCmd provides a command to list all inputs.
type InputListCmd struct {
	Input  bool `flag:"" help:"List all inputs."         aliases:"i"`
	Output bool `flag:"" help:"List all outputs."        aliases:"o"`
	Colour bool `flag:"" help:"List all colour sources." aliases:"c"`
	Ffmpeg bool `flag:"" help:"List all ffmpeg sources." aliases:"f"`
	Vlc    bool `flag:"" help:"List all VLC sources."    aliases:"v"`
	UUID   bool `flag:"" help:"Display UUIDs of inputs." aliases:"u"`
}

// Run executes the command to list all inputs.
func (cmd *InputListCmd) Run(ctx *context) error {
	resp, err := ctx.Client.Inputs.GetInputList(inputs.NewGetInputListParams())
	if err != nil {
		return err
	}

	t := table.New().Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ctx.Style.border))
	if cmd.UUID {
		t.Headers("Input Name", "Kind", "Muted", "UUID")
	} else {
		t.Headers("Input Name", "Kind", "Muted")
	}
	t.StyleFunc(func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 3)
		switch col {
		case 0:
			style = style.Align(lipgloss.Left)
		case 1:
			style = style.Align(lipgloss.Left)
		case 2:
			style = style.Align(lipgloss.Center)
		case 3:
			style = style.Align(lipgloss.Left)
		}
		switch {
		case row == table.HeaderRow:
			style = style.Bold(true).Align(lipgloss.Center)
		case row%2 == 0:
			style = style.Foreground(ctx.Style.evenRows)
		default:
			style = style.Foreground(ctx.Style.oddRows)
		}
		return style
	})

	sort.Slice(resp.Inputs, func(i, j int) bool {
		return resp.Inputs[i].InputName < resp.Inputs[j].InputName
	})

	for _, input := range resp.Inputs {
		var muteMark string
		resp, err := ctx.Client.Inputs.GetInputMute(
			inputs.NewGetInputMuteParams().WithInputName(input.InputName),
		)
		if err != nil {
			if err.Error() == "request GetInputMute: InvalidResourceState (604): The specified input does not support audio." {
				muteMark = "N/A"
			} else {
				return fmt.Errorf("failed to get input mute state: %w", err)
			}
		} else {
			muteMark = getEnabledMark(resp.InputMuted)
		}

		type filter struct {
			enabled bool
			keyword string
		}
		filters := []filter{
			{cmd.Input, "input"},
			{cmd.Output, "output"},
			{cmd.Colour, "color"}, // nolint: misspell
			{cmd.Ffmpeg, "ffmpeg"},
			{cmd.Vlc, "vlc"},
		}

		var added bool
		for _, f := range filters {
			if f.enabled && strings.Contains(input.InputKind, f.keyword) {
				if cmd.UUID {
					t.Row(input.InputName, input.InputKind, muteMark, input.InputUuid)
				} else {
					t.Row(input.InputName, input.InputKind, muteMark)
				}
				added = true
				break
			}
		}

		if !added && (!cmd.Input && !cmd.Output && !cmd.Colour && !cmd.Ffmpeg && !cmd.Vlc) {
			if cmd.UUID {
				t.Row(input.InputName, snakeCaseToTitleCase(input.InputKind), muteMark, input.InputUuid)
			} else {
				t.Row(input.InputName, snakeCaseToTitleCase(input.InputKind), muteMark)
			}
		}
	}
	fmt.Fprintln(ctx.Out, t.Render())
	return nil
}

// InputMuteCmd provides a command to mute an input.
type InputMuteCmd struct {
	InputName string `arg:"" help:"Name of the input to mute."`
}

// Run executes the command to mute an input.
func (cmd *InputMuteCmd) Run(ctx *context) error {
	_, err := ctx.Client.Inputs.SetInputMute(
		inputs.NewSetInputMuteParams().WithInputName(cmd.InputName).WithInputMuted(true),
	)
	if err != nil {
		return fmt.Errorf("failed to mute input: %w", err)
	}

	fmt.Fprintf(ctx.Out, "Muted input: %s\n", ctx.Style.Highlight(cmd.InputName))
	return nil
}

// InputUnmuteCmd provides a command to unmute an input.
type InputUnmuteCmd struct {
	InputName string `arg:"" help:"Name of the input to unmute."`
}

// Run executes the command to unmute an input.
func (cmd *InputUnmuteCmd) Run(ctx *context) error {
	_, err := ctx.Client.Inputs.SetInputMute(
		inputs.NewSetInputMuteParams().WithInputName(cmd.InputName).WithInputMuted(false),
	)
	if err != nil {
		return fmt.Errorf("failed to unmute input: %w", err)
	}

	fmt.Fprintf(ctx.Out, "Unmuted input: %s\n", ctx.Style.Highlight(cmd.InputName))
	return nil
}

// InputToggleCmd provides a command to toggle the mute state of an input.
type InputToggleCmd struct {
	InputName string `arg:"" help:"Name of the input to toggle."`
}

// Run executes the command to toggle the mute state of an input.
func (cmd *InputToggleCmd) Run(ctx *context) error {
	// Get the current mute state of the input
	resp, err := ctx.Client.Inputs.GetInputMute(
		inputs.NewGetInputMuteParams().WithInputName(cmd.InputName),
	)
	if err != nil {
		return fmt.Errorf("failed to get input mute state: %w", err)
	}
	// Toggle the mute state
	newMuteState := !resp.InputMuted
	_, err = ctx.Client.Inputs.SetInputMute(
		inputs.NewSetInputMuteParams().WithInputName(cmd.InputName).WithInputMuted(newMuteState),
	)
	if err != nil {
		return fmt.Errorf("failed to toggle input mute state: %w", err)
	}

	if newMuteState {
		fmt.Fprintf(ctx.Out, "Muted input: %s\n", ctx.Style.Highlight(cmd.InputName))
	} else {
		fmt.Fprintf(ctx.Out, "Unmuted input: %s\n", ctx.Style.Highlight(cmd.InputName))
	}
	return nil
}

// InputCreateCmd provides a command to create an input.
type InputCreateCmd struct {
	Kind   string `arg:"" help:"Input kind (e.g., coreaudio_input_capture, pulse_input_capture)." required:""`
	Name   string `arg:"" help:"Name for the input." required:""`
	Device string `arg:"" help:"Device ID or name." required:""`
}

// Run executes the command to create an input.
func (cmd *InputCreateCmd) Run(ctx *context) error {
	// Get current scene
	currentScene, err := ctx.Client.Scenes.GetCurrentProgramScene()
	if err != nil {
		return fmt.Errorf("failed to get current scene: %w", err)
	}
	
	settings := map[string]interface{}{
		"device_id": cmd.Device,
	}
	
	_, err = ctx.Client.Inputs.CreateInput(
		inputs.NewCreateInputParams().
			WithInputKind(cmd.Kind).
			WithInputName(cmd.Name).
			WithInputSettings(settings).
			WithSceneName(currentScene.CurrentProgramSceneName),
	)
	if err != nil {
		return fmt.Errorf("failed to create input: %w", err)
	}
	
	fmt.Fprintf(ctx.Out, "Created input: %s (%s) in scene %s\n", 
		ctx.Style.Highlight(cmd.Name), cmd.Kind, ctx.Style.Highlight(currentScene.CurrentProgramSceneName))
	return nil
}

// InputShowCmd provides a command to show input details.
type InputShowCmd struct {
	Name string `arg:"" help:"Name of the input to show." required:""`
}

// Run executes the command to show input details.
func (cmd *InputShowCmd) Run(ctx *context) error {
	// Get input settings
	settingsResp, err := ctx.Client.Inputs.GetInputSettings(
		inputs.NewGetInputSettingsParams().WithInputName(cmd.Name),
	)
	if err != nil {
		return fmt.Errorf("failed to get input settings: %w", err)
	}
	
	// Get input kind from input list
	listResp, err := ctx.Client.Inputs.GetInputList(inputs.NewGetInputListParams())
	if err != nil {
		return fmt.Errorf("failed to get input list: %w", err)
	}
	
	var inputKind string
	for _, input := range listResp.Inputs {
		if input.InputName == cmd.Name {
			inputKind = input.InputKind
			break
		}
	}
	
	fmt.Fprintf(ctx.Out, "Input Details:\n")
	fmt.Fprintf(ctx.Out, "  Name: %s\n", ctx.Style.Highlight(cmd.Name))
	fmt.Fprintf(ctx.Out, "  Kind: %s\n", inputKind)
	
	// Display device information if available
	if deviceID, ok := settingsResp.InputSettings["device_id"].(string); ok {
		fmt.Fprintf(ctx.Out, "  Device ID: %s\n", deviceID)
	}
	if deviceName, ok := settingsResp.InputSettings["device_name"].(string); ok {
		fmt.Fprintf(ctx.Out, "  Device Name: %s\n", deviceName)
	}
	
	// Display other settings
	fmt.Fprintf(ctx.Out, "\nSettings:\n")
	for key, value := range settingsResp.InputSettings {
		if key != "device_id" && key != "device_name" {
			fmt.Fprintf(ctx.Out, "  %s: %v\n", key, value)
		}
	}
	
	// Get and display available device options
	if inputKind != "" {
		fmt.Fprintf(ctx.Out, "\nAvailable Device Options:\n")
		
		// Try to get device list using GetInputPropertiesListPropertyItems
		deviceListResp, err := ctx.Client.Inputs.GetInputPropertiesListPropertyItems(
			inputs.NewGetInputPropertiesListPropertyItemsParams().
				WithInputName(cmd.Name).
				WithPropertyName("device_id"),
		)
		if err == nil && len(deviceListResp.PropertyItems) > 0 {
			fmt.Fprintf(ctx.Out, "  Property: device_id\n")
			for _, item := range deviceListResp.PropertyItems {
				// Skip empty items
				if item.ItemName != "" || item.ItemValue != "" {
					fmt.Fprintf(ctx.Out, "    %s: %s\n", item.ItemName, item.ItemValue)
				}
			}
		} else {
			// Fallback: try other common property names for devices
			deviceProps := []string{
				"device_id", "video_device_id", "audio_device_id", "monitor_id",
				"device", "video_device", "audio_device", "capture_device",
				"input_device", "source_device", "camera_device", "microphone_device",
			}
			found := false
			
			for _, propName := range deviceProps {
				deviceListResp, err := ctx.Client.Inputs.GetInputPropertiesListPropertyItems(
					inputs.NewGetInputPropertiesListPropertyItemsParams().
						WithInputName(cmd.Name).
						WithPropertyName(propName),
				)
				if err == nil && len(deviceListResp.PropertyItems) > 0 {
					fmt.Fprintf(ctx.Out, "  Property: %s\n", propName)
					for _, item := range deviceListResp.PropertyItems {
						// Skip empty items
						if item.ItemName != "" || item.ItemValue != "" {
							fmt.Fprintf(ctx.Out, "    %s: %s\n", item.ItemName, item.ItemValue)
						}
					}
					found = true
					break
				}
			}
			
			if !found {
				fmt.Fprintf(ctx.Out, "  No device enumeration available for this input type (%s)\n", inputKind)
			}
		}
	}
	
	return nil
}
