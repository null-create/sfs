package cmd

import (
	"log"
	"path/filepath"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
)

var (
	// pointer to client instance
	cl *client.Client

	// main client command
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "SFS Client Commands",
		Long: `
SFS Client Commands. 

Examples:
	sfs client --new        Create a new SFS client.
	sfs client --start      Start the SFS client
	sfs client --stop       Stop the SFS client
	sfs client --list       List root entries
	sfs client --sync       Sync local files with the SFS server
		`,
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}

	newClientCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new SFS client",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgs := client.ClientConfig()
			newUser := auth.NewUser(
				cfgs.User,
				cfgs.UserAlias,
				cfgs.Email,
				filepath.Join(cfgs.Root, cfgs.User),
				cfgs.IsAdmin,
			)
			newClient, err := client.NewClient(newUser)
			if err != nil {
				return err
			}
			cl = newClient
			return nil
		},
	}

	// child commands
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "client sync commands",
		Long: `
			SFS Client side synchronization command. 
			
			Use the flags associated with this command to configure and execute
			Synchronization operations from the client to the server.
		`,
	}

	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "push files to the sfs server",
		RunE: func(c *cobra.Command, args []string) error {
			return cl.Push()
		},
	}

	pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "pull files from the sfs server",
		RunE: func(c *cobra.Command, args []string) error {
			idx := cl.GetServerIdx()
			if idx == nil {
				log.Print("no sync index received from server")
				return nil
			}
			return cl.Pull(idx)
		},
	}
)

func init() {
	clientCmd.AddCommand(pushCmd)
	clientCmd.AddCommand(pullCmd)
	clientCmd.AddCommand(syncCmd)
	clientCmd.AddCommand(newClientCmd)

	rootCmd.AddCommand(clientCmd)
}
