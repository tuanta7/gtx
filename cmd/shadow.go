package cmd

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
	"github.com/tuanta7/gtx/internal/git"
	"github.com/tuanta7/gtx/internal/profile"
)

var (
	shadowMessage string
	shadowProfile string
	shadowName    string
	shadowEmail   string
)

// shadowCmd represents the shadow command
var shadowCmd = &cobra.Command{
	Use:   "shadow",
	Short: "Commit with a temporary profile for shadowing",
	Long:  `Shadow allows you to commit using a temporary identity (name and email).`,
}

var shadowCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit with a shadow profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if shadowMessage == "" {
			return fmt.Errorf("commit message is required")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		repo, err := git.OpenRepository(cwd)
		if err != nil {
			return err
		}

		var sig *object.Signature
		if shadowName != "" || shadowEmail != "" {
			if shadowName == "" || shadowEmail == "" {
				return fmt.Errorf("both --name and --email must be provided if using flags")
			}
			sig = &object.Signature{
				Name:  shadowName,
				Email: shadowEmail,
				When:  time.Now(),
			}
		} else if shadowProfile != "" {
			cfg, err := profile.LoadConfig()
			if err != nil {
				return err
			}
			p, ok := cfg.Get(shadowProfile)
			if !ok {
				return fmt.Errorf("profile %q not found", shadowProfile)
			}
			sig = &object.Signature{
				Name:  p.Name,
				Email: p.Email,
				When:  time.Now(),
			}
		} else {
			return fmt.Errorf("either --profile or both --name and --email must be provided")
		}

		hash, err := repo.Commit(shadowMessage, sig)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Committed as %s <%s>\n", sig.Name, sig.Email)
		fmt.Fprintf(cmd.OutOrStdout(), "Hash: %s\n", hash.String())
		return nil
	},
}

var profileCmdGroup = &cobra.Command{
	Use:   "profile",
	Short: "Manage shadow profiles",
}

var profileAddCmd = &cobra.Command{
	Use:   "add <id>",
	Short: "Add a new shadow profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if shadowName == "" || shadowEmail == "" {
			return fmt.Errorf("--name and --email are required")
		}

		cfg, err := profile.LoadConfig()
		if err != nil {
			return err
		}

		cfg.Set(id, profile.Profile{
			Name:  shadowName,
			Email: shadowEmail,
		})

		if err := profile.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Profile %q added.\n", id)
		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all shadow profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := profile.LoadConfig()
		if err != nil {
			return err
		}

		if len(cfg.Profiles) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No shadow profiles found.")
			return nil
		}

		ids := make([]string, 0, len(cfg.Profiles))
		for id := range cfg.Profiles {
			ids = append(ids, id)
		}
		sort.Strings(ids)

		for _, id := range ids {
			p := cfg.Profiles[id]
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s <%s>\n", id, p.Name, p.Email)
		}

		return nil
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a shadow profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		cfg, err := profile.LoadConfig()
		if err != nil {
			return err
		}

		if _, ok := cfg.Get(id); !ok {
			return fmt.Errorf("profile %q not found", id)
		}

		cfg.Delete(id)
		if err := profile.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Profile %q deleted.\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shadowCmd)

	shadowCmd.AddCommand(shadowCommitCmd)
	shadowCmd.AddCommand(profileCmdGroup)

	profileCmdGroup.AddCommand(profileAddCmd)
	profileCmdGroup.AddCommand(profileListCmd)
	profileCmdGroup.AddCommand(profileDeleteCmd)

	// Flags for commit
	shadowCommitCmd.Flags().StringVarP(&shadowMessage, "message", "m", "", "Commit message")
	shadowCommitCmd.Flags().StringVarP(&shadowProfile, "profile", "p", "", "Shadow profile ID")
	shadowCommitCmd.Flags().StringVar(&shadowName, "name", "", "Shadow author name")
	shadowCommitCmd.Flags().StringVar(&shadowEmail, "email", "", "Shadow author email")

	// Flags for profile add
	profileAddCmd.Flags().StringVar(&shadowName, "name", "", "Shadow author name")
	profileAddCmd.Flags().StringVar(&shadowEmail, "email", "", "Shadow author email")
}
