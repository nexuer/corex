package cli

import (
	"flag"
	"fmt"
	"os"

	internalruntime "github.com/nexuer/corex/internal/runtime"
	"github.com/nexuer/log"
	"github.com/nexuer/log/logmgr"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command) int {
	if err := run(cmd); err != nil {
		if !logmgr.Initialized() {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}

		defer logmgr.Close()

		log.Error(err)

		w := log.Default().Writer()
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
	return run(cmd)
}

func run(cmd *cobra.Command) error {
	parentName := cmd.Name()
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
	logmgr.AddFlags(fs)
	cmd.PersistentFlags().AddGoFlagSet(fs)
	fields := log.DefaultFields
	if cmd.Version != "" {
		fields = append(fields, log.String("version", cmd.Version))
	}
	// Inject logs.InitLogs after command line parsing into one of the
	// PersistentPre* functions.
	switch {
	case cmd.PersistentPreRun != nil:
		pre := cmd.PersistentPreRun
		cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			initLogs(parentName, cmd.Name(), fields...)
			pre(cmd, args)
		}
	case cmd.PersistentPreRunE != nil:
		pre := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			initLogs(parentName, cmd.Name(), fields...)
			return pre(cmd, args)
		}
	default:
		cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			initLogs(parentName, cmd.Name(), fields...)
		}
	}
	return cmd.Execute()
}

func initLogs(parentName string, subName string, fields ...log.Field) {
	name := parentName
	if parentName != subName {
		name += "_" + subName
	}
	opts := []logmgr.Option{
		logmgr.WithFields(fields...),
	}
	// init logmgr
	logmgr.Init(name, opts...)
	// reset kratos default logger
	internalruntime.SetDefaultLogger(log.Default())
}
