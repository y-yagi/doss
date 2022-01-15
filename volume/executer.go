package volume

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type Executer struct {
	cmd       *cobra.Command
	args      []string
	outStream io.Writer
}

func NewExecuter(cmd *cobra.Command, args []string, outStream io.Writer) (*Executer, error) {
	return &Executer{cmd: cmd, args: args, outStream: outStream}, nil
}

func (e *Executer) Execute() error {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		return err
	}

	volumes, err := cli.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		return err
	}

	list, err := e.cmd.Flags().GetBool("list")
	if err != nil {
		return err
	}

	if list {
		e.showList(volumes)
		return nil
	}

	pattern, err := e.cmd.Flags().GetString("find")
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, volume := range volumes.Volumes {
		wg.Add(1)
		go func(location string) {
			e.search(location, pattern)
			wg.Done()
		}(volume.Mountpoint)
	}
	wg.Wait()

	return nil
}

func (f *Executer) search(location string, name string) {
	err := filepath.Walk(location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		matched, _ := filepath.Match(name, info.Name())
		if matched {
			fmt.Fprintf(f.outStream, "%s\n", path)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(f.outStream, "%v\n", err)
	}
}

func (f *Executer) showList(volumes volume.VolumeListOKBody) {
	table := tablewriter.NewWriter(f.outStream)
	table.SetHeader([]string{"Driver", "Name", "Mountpoint"})

	for _, volume := range volumes.Volumes {
		table.Append([]string{volume.Driver, volume.Name, volume.Mountpoint})
	}

	table.Render()
}
