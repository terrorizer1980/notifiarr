package configfile

import (
	"os"
	"runtime"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	flag "github.com/spf13/pflag"
)

/* This file handles application cli flags. */

// Flags are our CLI input flags.
type Flags struct {
	*flag.FlagSet
	VerReq     bool
	Restart    bool
	Updated    bool
	PSlist     bool
	Write      string
	Curl       string
	ConfigFile string
	ExtraConf  []string
	EnvPrefix  string
	Headers    []string
}

// ParseArgs stores the cli flag data into the Flags pointer.
func (f *Flags) ParseArgs(args []string) {
	f.StringVarP(&f.ConfigFile, "config", "c",
		os.Getenv(mnd.DefaultEnvPrefix+"_CONFIG_FILE"), strings.Title(f.Name())+" Config File.")
	f.StringSliceVarP(&f.ExtraConf, "extraconfig", "e", nil, "This app supports multiple config files. "+
		"Separate with commas, or pass -e more than once.")
	f.StringVarP(&f.EnvPrefix, "prefix", "p", mnd.DefaultEnvPrefix, "Environment Variable Prefix.")
	f.BoolVarP(&f.VerReq, "version", "v", false, "Print the version and exit.")
	f.StringVar(&f.Curl, "curl", "", "GET a URL and display headers and payload.")
	f.StringSliceVar(&f.Headers, "header", nil, "Use with --curl to add a request header.")
	f.BoolVar(&f.PSlist, "ps", false, "Print the system process list; useful for 'process' service checks.")
	f.StringVarP(&f.Write, "write", "w", "", "Write new config file to provided path. Use - to overwrite '--config' file.")

	if runtime.GOOS == mnd.Windows {
		f.BoolVar(&f.Restart, "restart", false, "This is used by auto-update, do not call it.")
		f.BoolVar(&f.Updated, "updated", false, "This flag causes the app to print an 'updated' message.")
	}

	f.Parse(args) // nolint: errcheck
}
