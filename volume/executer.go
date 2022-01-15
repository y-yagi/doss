package volume

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/manifoldco/promptui"
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

	volumeOKBody, err := cli.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		return err
	}

	list, err := e.cmd.Flags().GetBool("list")
	if err != nil {
		return err
	}

	if list {
		e.showList(volumeOKBody.Volumes)
		return nil
	}

	remove, err := e.cmd.Flags().GetBool("remove")
	if err != nil {
		return err
	}

	if remove {
		return e.remove(cli, volumeOKBody.Volumes)
	}

	pattern, err := e.cmd.Flags().GetString("find")
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, volume := range volumeOKBody.Volumes {
		wg.Add(1)
		go func(location string) {
			e.search(location, pattern)
			wg.Done()
		}(volume.Mountpoint)
	}
	wg.Wait()

	return nil
}

func (e *Executer) search(location string, name string) {
	err := filepath.Walk(location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		matched, _ := filepath.Match(name, info.Name())
		if matched {
			fmt.Fprintf(e.outStream, "%s\n", path)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(e.outStream, "%v\n", err)
	}
}

func (e *Executer) showList(volumes []*types.Volume) {
	table := tablewriter.NewWriter(e.outStream)
	table.SetHeader([]string{"Driver", "Name", "Mountpoint"})

	for _, volume := range volumes {
		table.Append([]string{volume.Driver, volume.Name, volume.Mountpoint})
	}

	table.Render()
}

func (e *Executer) remove(client *client.Client, volumes []*types.Volume) error {
	items := []string{}
	idMap := map[string]string{}
	for _, volume := range volumes {
		key := fmt.Sprintf("%s - %s", volume.Driver, volume.Name)
		items = append(items, key)
		idMap[key] = volume.Name
	}

	prompt := promptui.Select{
		Label: "Select volume",
		Items: items,
		Size:  20,
	}

	_, item, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil
		}

		return fmt.Errorf("prompt failed %v", err)
	}

	id := idMap[item]
	if err = client.VolumeRemove(context.Background(), id, false); err != nil {
		return err
	}

	fmt.Fprintf(e.outStream, "Remove %s\n", item)
	return nil
}
