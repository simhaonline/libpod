package containers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/containers/libpod/cmd/podmanV2/registry"
	"github.com/containers/libpod/pkg/domain/entities"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	commitDescription = `Create an image from a container's changes. Optionally tag the image created, set the author with the --author flag, set the commit message with the --message flag, and make changes to the instructions with the --change flag.`

	commitCommand = &cobra.Command{
		Use:     "commit [flags] CONTAINER [IMAGE]",
		Short:   "Create new image based on the changed container",
		Long:    commitDescription,
		RunE:    commit,
		PreRunE: preRunE,
		Args:    cobra.MinimumNArgs(1),
		Example: `podman commit -q --message "committing container to image" reverent_golick image-committed
  podman commit -q --author "firstName lastName" reverent_golick image-committed
  podman commit -q --pause=false containerID image-committed
  podman commit containerID`,
	}

	// ChangeCmds is the list of valid Changes commands to passed to the Commit call
	ChangeCmds = []string{"CMD", "ENTRYPOINT", "ENV", "EXPOSE", "LABEL", "ONBUILD", "STOPSIGNAL", "USER", "VOLUME", "WORKDIR"}
)

var (
	commitOptions = entities.CommitOptions{
		ImageName: "",
	}
	iidFile string
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Mode:    []entities.EngineMode{entities.ABIMode, entities.TunnelMode},
		Command: commitCommand,
	})
	flags := commitCommand.Flags()
	flags.StringArrayVarP(&commitOptions.Changes, "change", "c", []string{}, "Apply the following possible instructions to the created image (default []): "+strings.Join(ChangeCmds, " | "))
	flags.StringVarP(&commitOptions.Format, "format", "f", "oci", "`Format` of the image manifest and metadata")
	flags.StringVarP(&iidFile, "iidfile", "", "", "`file` to write the image ID to")
	flags.StringVarP(&commitOptions.Message, "message", "m", "", "Set commit message for imported image")
	flags.StringVarP(&commitOptions.Author, "author", "a", "", "Set the author for the image committed")
	flags.BoolVarP(&commitOptions.Pause, "pause", "p", false, "Pause container during commit")
	flags.BoolVarP(&commitOptions.Quiet, "quiet", "q", false, "Suppress output")
	flags.BoolVar(&commitOptions.IncludeVolumes, "include-volumes", false, "Include container volumes as image volumes")

}
func commit(cmd *cobra.Command, args []string) error {
	container := args[0]
	if len(args) > 1 {
		commitOptions.ImageName = args[1]
	}
	if !commitOptions.Quiet {
		commitOptions.Writer = os.Stderr
	}

	response, err := registry.ContainerEngine().ContainerCommit(context.Background(), container, commitOptions)
	if err != nil {
		return err
	}
	if len(iidFile) > 0 {
		if err = ioutil.WriteFile(iidFile, []byte(response.Id), 0644); err != nil {
			return errors.Wrapf(err, "failed to write image ID to file %q", iidFile)
		}
	}
	fmt.Println(response.Id)
	return nil
}
