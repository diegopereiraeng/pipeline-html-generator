package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

var build = "1" // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "Harness-Pipeline-Status-Reporter"
	app.Usage = "CLI tool to report pipeline status to Harness"
	app.Action = run
	app.Version = fmt.Sprintf("1.0.%s", build)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "acc_id",
			Usage:  "e.g: 2_gVHyo9Qiu4dXvj-AcbC",
			EnvVar: "HARNESS_ACCOUNT_ID, ACCOUNT_ID, PLUGIN_ACCID",
		},
		cli.StringFlag{
			Name:   "org_id",
			Usage:  "default",
			EnvVar: "HARNESS_ORG_ID, PLUGIN_ORG_ID",
		},
		cli.StringFlag{
			Name:   "project_id",
			Usage:  "GIT_FLOW_DEMO",
			EnvVar: "HARNESS_PROJECT_ID, PLUGIN_PROJECT_ID",
		},
		cli.StringFlag{
			Name:   "pipeline_id",
			Usage:  "FAST_CISTO_SonarQube_Quality_Gate_Plugin",
			EnvVar: "HARNESS_PIPELINE_ID, PLUGIN_PIPELINE_ID",
		},
		cli.StringSliceFlag{
			Name:   "status_list",
			Usage:  "Comma-separated list of statuses to filter by. E.g: Success,Aborted",
			Value:  &cli.StringSlice{"Success", "Running"},
			EnvVar: "PLUGIN_STATUS_LIST",
		},
		cli.StringFlag{
			Name:   "repo_name",
			Usage:  "sonarqube-scanner",
			EnvVar: "DRONE_REPO_NAME, PLUGIN_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "branch",
			Usage:  "main",
			EnvVar: "CI_COMMIT_BRANCH, PLUGIN_BRANCH",
		},
		cli.StringFlag{
			Name:   "service_name",
			Usage:  "sonarqube-scanner",
			EnvVar: "PLUGIN_SERVICE_NAME",
		},
		cli.StringFlag{
			Name:   "harness_secret",
			Usage:  "Harness access token with visualization permissions",
			EnvVar: "PLUGIN_HARNESS_SECRET",
		},
		cli.StringFlag{
			Name:   "harness_pipe_execution_url",
			Usage:  "Provide a custom Harness execution URL, or it gonna take the current pipeline execution URL",
			EnvVar: "CI_BUILD_LINK, PLUGIN_HARNESS_PIPE_EXECUTION_URL",
		},
	}
	app.Run(os.Args)
}

func run(c *cli.Context) {
	if c.String("json_file_name") != "" && c.String("json_content") != "" {
		fmt.Println("Error: Please specify either json_file_name or json_content, but not both.")
		os.Exit(1)
	}

	config := Config{
		AccID:         c.String("acc_id"),
		OrgID:         c.String("org_id"),
		ProjectID:     c.String("project_id"),
		PipelineID:    c.String("pipeline_id"),
		StatusList:    c.StringSlice("status_list"),
		RepoName:      c.String("repo_name"),
		Branch:        c.String("branch"),
		ServiceName:   c.String("service_name"),
		HarnessSecret: c.String("harness_secret"),
	}

	plugin := Plugin{Config: config}
	if err := plugin.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
