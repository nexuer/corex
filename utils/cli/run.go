package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/nexuer/corex/log"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command) int {
	if logsInitialized, err := run(cmd); err != nil {
		if !logsInitialized {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}

		defer log.Close()

		log.Error(err)

		w := log.Writer()
		if w != os.Stderr && w != os.Stdout {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		return 1
	}
	return 0
}

// RunNoErrOutput is a version of Run which returns the cobra command error
// instead of printing it.
func RunNoErrOutput(cmd *cobra.Command) error {
	_, err := run(cmd)
	return err
}

func run(cmd *cobra.Command) (logsInitialized bool, err error) {
	cmdName := cmd.Name()
	if !cmd.SilenceUsage {
		cmd.SilenceUsage = true
		cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
			// Re-enable usage printing.
			c.SilenceUsage = false
			return err
		})
	}

	// In all cases error printing is done below.
	cmd.SilenceErrors = true

	// This is idempotent.
	fs := flag.NewFlagSet("logging", flag.ContinueOnError)
	log.AddFlags(fs)
	cmd.PersistentFlags().AddGoFlagSet(fs)

	// Inject logs.InitLogs after command line parsing into one of the
	// PersistentPre* functions.
	switch {
	case cmd.PersistentPreRun != nil:
		pre := cmd.PersistentPreRun
		cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			initLogs(cmdName, cmd.Name())
			logsInitialized = true
			pre(cmd, args)
		}
	case cmd.PersistentPreRunE != nil:
		pre := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			initLogs(cmdName, cmd.Name())
			logsInitialized = true
			return pre(cmd, args)
		}
	default:
		cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			initLogs(cmdName, cmd.Name())
			logsInitialized = true
		}
	}
	err = cmd.Execute()
	return
}

func initLogs(parentName string, subName string) {
	name := parentName
	if parentName != subName {
		name += "_" + subName
	}
	log.InitManager(name, "ts", log.DefaultTimestamp, "caller", log.Caller(7))
}
