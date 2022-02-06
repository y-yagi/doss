package volume

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

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
	client    *client.Client
}

const (
	ListFlagName   = "list"
	RemoveFlagName = "remove"
	FindFlagName   = "find"
)

func NewExecuter(cmd *cobra.Command, args []string, outStream io.Writer) (*Executer, error) {
	return &Executer{cmd: cmd, args: args, outStream: outStream}, nil
}

func (e *Executer) Execute() error {
	var err error
	e.client, err = client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		return err
	}

	list, err := e.cmd.Flags().GetBool(ListFlagName)
	if err != nil {
		return err
	}

	if list {
		return e.showList()
	}

	remove, err := e.cmd.Flags().GetBool(RemoveFlagName)
	if err != nil {
		return err
	}

	if remove {
		return e.remove()
	}

	pattern, err := e.cmd.Flags().GetString(FindFlagName)
	if err != nil {
		return err
	}

	if len(pattern) == 0 {
		return e.cmd.Usage()
	}

	return e.search(pattern)
}

func (e *Executer) search(pattern string) error {
	volumeOKBody, err := e.client.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, volume := range volumeOKBody.Volumes {
		wg.Add(1)
		go func(location string) {
			e.walk(location, pattern)
			wg.Done()
		}(volume.Mountpoint)
	}
	wg.Wait()

	return nil
}

func (e *Executer) walk(location string, name string) {
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

func (e *Executer) showList() error {
	volumeOKBody, err := e.client.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(e.outStream)
	table.SetHeader([]string{"Driver", "Name", "Mountpoint"})

	for _, volume := range volumeOKBody.Volumes {
		table.Append([]string{volume.Driver, volume.Name, volume.Mountpoint})
	}

	table.Render()
	return nil
}

func (e *Executer) remove() error {
	volumeOKBody, err := e.client.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		return err
	}

	items := []string{}
	idMap := map[string]string{}
	for _, volume := range volumeOKBody.Volumes {
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
	if err = e.client.VolumeRemove(context.Background(), id, false); err != nil {
		if strings.Contains(err.Error(), "volume is in use") {
			containerNames, newerr := e.getContainerNames(err)
			if newerr == nil {
				return fmt.Errorf("%w\n%s", err, containerNames)
			}
		}
		return err
	}

	fmt.Fprintf(e.outStream, "Remove %s\n", item)
	return nil
}

func (e *Executer) getContainerNames(err error) (string, error) {
	names := ""
	ids := e.getContainerIDsFromError(err)
	for _, id := range ids {
		container, err := e.client.ContainerInspect(context.Background(), id)
		if err != nil {
			return "", err
		}
		key := fmt.Sprintf("[%s - %s (%s) ]", container.Image, container.Name, container.ID)
		names += "," + key
	}

	return names, nil
}

func (e *Executer) getContainerIDsFromError(err error) []string {
	re := regexp.MustCompile(`\[(.*)\]`)
	ids := strings.Trim(string(re.Find([]byte(err.Error()))), "[]")
	return strings.Split(ids, ",")
}
