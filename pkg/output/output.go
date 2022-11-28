package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"k8s.io/utils/strings/slices"
)

type OutputType string

const (
	OUTPUT_TABLE OutputType = "table"
	OUTPUT_JSON  OutputType = "json"
	OUTPUT_YAML  OutputType = "yaml"
)

type TableOutput struct {
	Headers       []string
	SubHeaders    []string
	Data          [][]interface{}
	ColumnConfigs []table.ColumnConfig
}

type Output struct {
	Table      TableOutput
	DataOutput interface{}
}

func (o *Output) AddSubHeaders(headers []string) {
	o.Table.SubHeaders = headers
}

var (
	OutputFormat string
)

func (o *Output) VariableOutput() error {

	if !slices.Contains([]string{string(OUTPUT_TABLE), string(OUTPUT_JSON), string(OUTPUT_YAML)}, OutputFormat) {
		return fmt.Errorf("%s is not a valid output format", OutputFormat)
	}

	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}

	if OutputFormat == "table" {
		t := table.NewWriter()
		t.SetStyle(table.StyleLight)

		t.Style().Options.DrawBorder = false
		t.Style().Options.SeparateColumns = false
		t.Style().Options.SeparateHeader = false

		t.SetOutputMirror(os.Stdout)

		tHeaders := make(table.Row, len(o.Table.Headers))
		for i, header := range o.Table.Headers {
			tHeaders[i] = header
		}
		t.AppendHeader(tHeaders, rowConfigAutoMerge)

		if len(o.Table.SubHeaders) > 0 {
			subHeaders := make(table.Row, len(o.Table.SubHeaders))
			for i, header := range o.Table.SubHeaders {
				subHeaders[i] = header
			}
			t.AppendHeader(subHeaders, rowConfigAutoMerge)
		}

		if len(o.Table.ColumnConfigs) > 0 {
			t.SetColumnConfigs(o.Table.ColumnConfigs)
		}

		if len(o.Table.Data) > 0 {
			for _, c := range o.Table.Data {
				t.AppendRow(c, rowConfigAutoMerge)
			}
		}

		t.Render()
	} else if OutputFormat == "json" {
		jsonOutput, err := json.MarshalIndent(o.DataOutput, "", "  ")

		if err != nil {
			return err
		}

		fmt.Println(string(jsonOutput))
	} else if OutputFormat == "yaml" {
		yamlOutput, err := yaml.Marshal(o.DataOutput)

		if err != nil {
			return err
		}

		fmt.Println(string(yamlOutput))
	}

	return nil
}

func NewOutput(table TableOutput, dataOutput interface{}) *Output {

	return &Output{table, dataOutput}
}

func Confirmation(prompt string) error {
	yes := viper.GetBool("yes")

	if !yes {
		prompt := promptui.Prompt{
			Label:     prompt,
			IsConfirm: true,
		}

		_, err := prompt.Run()

		if err != nil {
			return fmt.Errorf("User cancelled action")
		}
	}

	return nil
}
