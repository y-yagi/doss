package attach

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/y-yagi/goext/arr"
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

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		fmt.Fprint(e.outStream, "No containers\n")
		return nil
	}

	items := []string{}
	idMap := map[string]string{}
	for _, container := range containers {
		key := fmt.Sprintf("%s - %s", container.Image, arr.Join(container.Names, ","))
		items = append(items, key)
		idMap[key] = container.ID
	}

	prompt := promptui.Select{
		Label: "Select container",
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
	fmt.Fprintf(e.outStream, "Attach to %s\n", item)

	cmd := exec.Command("docker", "attach", id)
	cmd.Stdout = e.outStream
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
