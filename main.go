package main

import (
	"encoding/json"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var revision string // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "terraform plugin"
	app.Usage = "terraform plugin"
	app.Action = run
	app.Version = revision
	app.Flags = []cli.Flag{

		//
		// plugin args
		//

		cli.StringSliceFlag{
			Name:   "actions",
			Usage:  "a list of actions to have terraform perform",
			EnvVar: "PLUGIN_ACTIONS",
			Value:  &cli.StringSlice{"validate", "plan", "apply"},
		},
		cli.StringFlag{
			Name:   "ca_cert",
			Usage:  "ca cert to add to your environment to allow terraform to use internal/private resources",
			EnvVar: "PLUGIN_CA_CERT",
		},
		cli.StringFlag{
			Name:   "env_file",
			Usage:  "pass filename to source it and load variables into current shell",
			EnvVar: "PLUGIN_ENV_FILE",
		},
		cli.StringFlag{
			Name:   "init_options",
			Usage:  "options for the init command. See https://www.terraform.io/docs/commands/init.html",
			EnvVar: "PLUGIN_INIT_OPTIONS",
		},
		cli.StringFlag{
			Name:   "summarize_options",
			Usage:  "options for the tf-summarize command. See https://github.com/dineshba/tf-summarize#usage",
			EnvVar: "PLUGIN_SUMMARIZE_OPTIONS",
		},
		cli.StringFlag{
			Name:   "fmt_options",
			Usage:  "options for the fmt command. See https://www.terraform.io/docs/commands/fmt.html",
			EnvVar: "PLUGIN_FMT_OPTIONS",
		},
		cli.IntFlag{
			Name:   "parallelism",
			Usage:  "The number of concurrent operations as Terraform walks its graph",
			EnvVar: "PLUGIN_PARALLELISM",
		},
		cli.BoolFlag{
			Name:   "skip_init",
			Usage:  "skip terraform init (useful for usage with s3-cache)",
			EnvVar: "PLUGIN_SKIP_INIT",
		},
		cli.BoolFlag{
			Name:   "skip_cleanup",
			Usage:  "skip removal of .terraform/ (useful for usage with s3-cache)",
			EnvVar: "PLUGIN_SKIP_CLEANUP",
		},
		cli.StringFlag{
			Name:   "netrc.machine",
			Usage:  "netrc machine",
			EnvVar: "DRONE_NETRC_MACHINE",
		},
		cli.StringFlag{
			Name:   "netrc.username",
			Usage:  "netrc username",
			EnvVar: "DRONE_NETRC_USERNAME",
		},
		cli.StringFlag{
			Name:   "netrc.password",
			Usage:  "netrc password",
			EnvVar: "DRONE_NETRC_PASSWORD",
		},
		cli.StringFlag{
			Name:   "role_arn_to_assume",
			Usage:  "A role to assume before running the terraform commands",
			EnvVar: "PLUGIN_ROLE_ARN_TO_ASSUME",
		},
		cli.StringFlag{
			Name:   "root_dir",
			Usage:  "The root directory where the terraform files live. When unset, the top level directory will be assumed",
			EnvVar: "PLUGIN_ROOT_DIR",
		},
		cli.StringFlag{
			Name:   "secrets",
			Usage:  "a map of secrets to pass to the Terraform `plan` and `apply` commands. Each value is passed as a `<key>=<ENV>` option",
			EnvVar: "PLUGIN_SECRETS",
		},
		cli.BoolFlag{
			Name:   "sensitive",
			Usage:  "whether or not to suppress terraform commands to stdout",
			EnvVar: "PLUGIN_SENSITIVE",
		},
		cli.StringSliceFlag{
			Name:   "targets",
			Usage:  "targets to run apply or plan on",
			EnvVar: "PLUGIN_TARGETS",
		},
		cli.StringFlag{
			Name:   "tf.version",
			Usage:  "terraform version to use",
			EnvVar: "PLUGIN_TF_VERSION",
		},
		cli.StringFlag{
			Name:   "vars",
			Usage:  "a map of variables to pass to the Terraform `plan` and `apply` commands. Each value is passed as a `<key>=<value>` option",
			EnvVar: "PLUGIN_VARS",
		},
		cli.StringSliceFlag{
			Name:   "var_files",
			Usage:  "a list of var files to use. Each value is passed as -var-file=<value>",
			EnvVar: "PLUGIN_VAR_FILES",
		},
		cli.StringFlag{
			Name:   "tf_data_dir",
			Usage:  "changes the location where Terraform keeps its per-working-directory data, such as the current remote backend configuration",
			EnvVar: "PLUGIN_TF_DATA_DIR",
		},
		cli.BoolFlag{
			Name:   "disable_refresh",
			Usage:  "whether or not to disable refreshing state before `plan` and `apply` commands",
			EnvVar: "PLUGIN_DISABLE_REFRESH",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	logrus.WithFields(logrus.Fields{
		"Revision": revision,
	}).Info("Drone Terraform Plugin Version")

	if c.String("env_file") != "" {
		_ = godotenv.Load(c.String("env_file"))
	}

	var vars map[string]string
	if c.String("vars") != "" {
		if err := json.Unmarshal([]byte(c.String("vars")), &vars); err != nil {
			panic(err)
		}
	}
	var secrets map[string]string
	if c.String("secrets") != "" {
		if err := json.Unmarshal([]byte(c.String("secrets")), &secrets); err != nil {
			panic(err)
		}
	}

	initOptions := InitOptions{}
	json.Unmarshal([]byte(c.String("init_options")), &initOptions)
	fmtOptions := FmtOptions{}
	json.Unmarshal([]byte(c.String("fmt_options")), &fmtOptions)
	summarizeOptions := SummarizeOptions{}
	json.Unmarshal([]byte(c.String("summarize_options")), &summarizeOptions)

	plugin := Plugin{
		Config: Config{
			Actions:          c.StringSlice("actions"),
			Vars:             vars,
			Secrets:          secrets,
			InitOptions:      initOptions,
			FmtOptions:       fmtOptions,
			SummarizeOptions: summarizeOptions,
			Cacert:           c.String("ca_cert"),
			Sensitive:        c.Bool("sensitive"),
			RoleARN:          c.String("role_arn_to_assume"),
			RootDir:          c.String("root_dir"),
			SkipInit:         c.Bool("skip_init"),
			SkipCleanup:      c.Bool("skip_cleanup"),
			Parallelism:      c.Int("parallelism"),
			Targets:          c.StringSlice("targets"),
			VarFiles:         c.StringSlice("var_files"),
			TerraformDataDir: c.String("tf_data_dir"),
			DisableRefresh:   c.Bool("disable_refresh"),
		},
		Netrc: Netrc{
			Login:    c.String("netrc.username"),
			Machine:  c.String("netrc.machine"),
			Password: c.String("netrc.password"),
		},
		Terraform: Terraform{
			Version: c.String("tf.version"),
		},
	}

	return plugin.Exec()
}
