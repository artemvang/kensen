package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/artemvang/kensen/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "kensen",
	Short: "Kensen is a simple database migration tool",
	Args:  cobra.NoArgs,
}

var initCmd = &cobra.Command{
	Use:          "init",
	Short:        "Initialize migration table",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		k, err := kensen.Create(dbURI, migrationsDir)
		if err != nil {
			return err
		}

		if err := k.Init(); err != nil {
			return err
		}
		fmt.Printf("Successfully initialized migration table")
		return nil
	},
}

var newCmd = &cobra.Command{
	Use:          "new [migration name]",
	Short:        "Create new migration file",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		k, err := kensen.Create(dbURI, migrationsDir)
		if err != nil {
			return err
		}

		migration, err := k.New(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Successfully created migration %s table", *migration)
		return nil
	},
}

var applyCmd = &cobra.Command{
	Use:          "apply",
	Short:        "Apply migrations",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		k, err := kensen.Create(dbURI, migrationsDir)
		if err != nil {
			return err
		}

		migrationStatuses, err := k.Apply()
		if err != nil {
			return err
		}
		for _, mig := range migrationStatuses {
			fmt.Printf("%s - %s", mig.Migration, mig.Status.String())
			return mig.Err
		}

		return nil
	},
}

var dbURI string
var cfgFile string
var migrationsDir string

func init() {
	cobra.OnInitialize(func() {
		viper.SetConfigFile(cfgFile)

		viper.SetEnvPrefix("KS")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}

		postInitCommands(rootCmd.Commands())
	})

	flags := rootCmd.PersistentFlags()

	flags.StringVarP(&cfgFile, "config", "f", ".kensen.yaml", "Config file")
	flags.StringVar(&dbURI, "db-uri", "", "Database URI (e.g. file:///example.db, postgresql://user:pass@localhost/db)")
	flags.StringVar(&migrationsDir, "migrations-dir", "./migrations", "Migrations directory")
	rootCmd.MarkPersistentFlagFilename("config", "yaml")

	rootCmd.AddCommand(initCmd, newCmd, applyCmd)
}

func postInitCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		presetRequiredFlags(cmd)
		if cmd.HasSubCommands() {
			postInitCommands(cmd.Commands())
		}
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
	viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func exitError(err error) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s", err.Error())
	os.Exit(1)
}
