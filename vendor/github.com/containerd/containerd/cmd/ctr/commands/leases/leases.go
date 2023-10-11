/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package leases

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/containerd/containerd/cmd/ctr/commands"
	"github.com/containerd/containerd/leases"
	"github.com/urfave/cli"
)

// Command is the cli command for managing content
var Command = cli.Command{
	Name:  "leases",
	Usage: "Manage leases",
	Subcommands: cli.Commands{
		listCommand,
		createCommand,
		deleteCommand,
	},
}

var listCommand = cli.Command{

	Name:        "list",
	Aliases:     []string{"ls"},
	Usage:       "List all active leases",
	ArgsUsage:   "[flags] <filter>",
	Description: "list active leases by containerd",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet, q",
			Usage: "Print only the blob digest",
		},
	},
	Action: func(context *cli.Context) error {
		var (
			filters = context.Args()
			quiet   = context.Bool("quiet")
		)
		client, ctx, cancel, err := commands.NewClient(context)
		if err != nil {
			return err
		}
		defer cancel()

		ls := client.LeasesService()

		leaseList, err := ls.List(ctx, filters...)
		if err != nil {
			return fmt.Errorf("failed to list leases: %w", err)
		}
		if quiet {
			for _, l := range leaseList {
				fmt.Println(l.ID)
			}
			return nil
		}
		tw := tabwriter.NewWriter(os.Stdout, 1, 8, 1, ' ', 0)
		fmt.Fprintln(tw, "ID\tCREATED AT\tLABELS\t")
		for _, l := range leaseList {
			labels := "-"
			if len(l.Labels) > 0 {
				var pairs []string
				for k, v := range l.Labels {
					pairs = append(pairs, fmt.Sprintf("%v=%v", k, v))
				}
				sort.Strings(pairs)
				labels = strings.Join(pairs, ",")
			}

			fmt.Fprintf(tw, "%v\t%v\t%s\t\n",
				l.ID,
				l.CreatedAt.Local().Format(time.RFC3339),
				labels)
		}

		return tw.Flush()
	},
}

var createCommand = cli.Command{
	Name:        "create",
	Usage:       "Create lease",
	ArgsUsage:   "[flags] <label>=<value> ...",
	Description: "create a new lease",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Usage: "Set the id for the lease, will be generated by default",
		},
		cli.DurationFlag{
			Name:  "expires, x",
			Usage: "Expiration of lease (0 value will not expire)",
			Value: 24 * time.Hour,
		},
	},
	Action: func(context *cli.Context) error {
		var labelstr = context.Args()
		client, ctx, cancel, err := commands.NewClient(context)
		if err != nil {
			return err
		}
		defer cancel()

		ls := client.LeasesService()
		opts := []leases.Opt{}
		if len(labelstr) > 0 {
			labels := map[string]string{}
			for _, lstr := range labelstr {
				k, v, _ := strings.Cut(lstr, "=")
				labels[k] = v
			}
			opts = append(opts, leases.WithLabels(labels))
		}

		if id := context.String("id"); id != "" {
			opts = append(opts, leases.WithID(id))
		}
		if exp := context.Duration("expires"); exp > 0 {
			opts = append(opts, leases.WithExpiration(exp))
		}

		l, err := ls.Create(ctx, opts...)
		if err != nil {
			return err
		}

		fmt.Println(l.ID)

		return nil
	},
}

var deleteCommand = cli.Command{
	Name:        "delete",
	Aliases:     []string{"del", "remove", "rm"},
	Usage:       "Delete a lease",
	ArgsUsage:   "[flags] <lease id> ...",
	Description: "delete a lease",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "sync",
			Usage: "Synchronously remove leases and all unreferenced resources",
		},
	},
	Action: func(context *cli.Context) error {
		var lids = context.Args()
		if len(lids) == 0 {
			return cli.ShowSubcommandHelp(context)
		}
		client, ctx, cancel, err := commands.NewClient(context)
		if err != nil {
			return err
		}
		defer cancel()

		ls := client.LeasesService()
		sync := context.Bool("sync")
		for i, lid := range lids {
			var opts []leases.DeleteOpt
			if sync && i == len(lids)-1 {
				opts = append(opts, leases.SynchronousDelete)
			}

			lease := leases.Lease{
				ID: lid,
			}
			if err := ls.Delete(ctx, lease, opts...); err != nil {
				return err
			}
			fmt.Println(lid)
		}

		return nil
	},
}
