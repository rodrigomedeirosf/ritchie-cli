package cmd

import (
	"fmt"

	"github.com/ZupIT/ritchie-cli/pkg/formula/repo"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

// listRepoCmd type for list repo command
type listRepoCmd struct {
	repo.Lister
}

// NewListRepoCmd creates a new cmd instance
func NewListRepoCmd(ls repo.Lister) *cobra.Command {
	l := &listRepoCmd{ls}

	cmd := &cobra.Command{
		Use:     "repo",
		Short:   "List all repositories.",
		Example: "rit list repo",
		RunE:    l.runFunc(),
	}

	return cmd
}

func (l listRepoCmd) runFunc() CommandRunnerFunc {
	return func(cmd *cobra.Command, args []string) error {
		rr, err := l.List()
		if err != nil {
			return err
		}

		printList(rr)

		return nil
	}
}

func printList(rr []repo.Repository) {
	table := uitable.New()
	table.AddRow("NAME", "URL")
	for _, re := range rr {
		table.AddRow(re.Name, re.TreePath)
	}
	raw := table.Bytes()
	raw = append(raw, []byte("\n")...)
	fmt.Println(string(raw))
}
