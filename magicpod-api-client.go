package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Magic-Pod/magicpod-api-client/common"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.99.39.2"
	app.Name = "magicpod-api-client"
	app.Usage = "Simple and useful wrapper for MagicPod Web API"
	app.Flags = []cli.Flag{
		// hidden option only for MagicPod developers
		cli.StringFlag{
			Name:   "url-base",
			Value:  "https://app.magicpod.com",
			Hidden: true,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "batch-run",
			Usage: "Run batch test",
			Flags: append(commonFlags(), []cli.Flag{
				cli.IntFlag{
					Name:  "test_settings_number, S",
					Usage: "Test settings number defined in the project batch run page",
				},
				cli.StringFlag{
					Name:  "setting, s",
					Usage: "Test setting in JSON format. Please check https://app.magicpod.com/api/v1.0/doc/ for more detail",
				},
				cli.BoolFlag{
					Name:  "no_wait, n",
					Usage: "Return immediately without waiting the batch run to be finished",
				},
				cli.IntFlag{
					Name:  "wait_limit, w",
					Usage: "Wait limit in seconds. If 0 is specified, the value is test count x 10 minutes",
				},
			}...),
			Action: batchRunAction,
		},
		{
			Name:  "get-batch-run",
			Usage: "Get batch run result",
			Flags: append(commonFlags(), []cli.Flag{
				cli.IntFlag{
					Name:  "batch_run_number, b",
					Usage: "Batch run number",
				},
			}...),
			Action: getBatchRunAction,
		},
		{
			Name:  "get-batch-runs",
			Usage: "Get the batch runs information in the **most recent first** order.",
			Flags: append(commonFlags(), []cli.Flag{
				cli.IntFlag{
					Name:  "count, c",
					Usage: "The maximum number of records to retrieve.",
					Value: 20,
				},
				cli.IntFlag{
					Name:     "max_batch_run_number, max",
					Usage:    "The most recent batch run number to start retrieving records from.",
					Required: false,
				},
				cli.IntFlag{
					Name:     "min_batch_run_number, min",
					Usage:    "The least recent batch run number to stop retrieving records at.",
					Required: false,
				},
			}...),
			Action: getBatchRunsAction,
		},
		{
			Name:   "latest-batch-run-no",
			Usage:  "Get the latest batch run number",
			Flags:  commonFlags(),
			Action: latestBatchRunNoAction,
		},
		{
			Name:  "upload-app",
			Usage: "Upload app/ipa/apk file",
			Flags: append(commonFlags(), []cli.Flag{
				cli.StringFlag{
					Name:  "app_path, a",
					Usage: "Path to the app/ipa/apk file to upload",
				},
			}...),
			Action: uploadAppAction,
		},
		{
			Name:  "delete-app",
			Usage: "Delete uploaded app/ipa/apk file",
			Flags: append(commonFlags(), []cli.Flag{
				cli.IntFlag{
					Name:  "app_file_number, a",
					Usage: "File number of the uploaded file",
				},
			}...),
			Action: deleteAppAction,
		},
		{
			Name:  "get-screenshots",
			Usage: "Download screenshots for a batch run",
			Flags: append(commonFlags(), []cli.Flag{
				cli.IntFlag{
					Name:  "batch_run_number, b",
					Usage: "Batch run number",
				},
				cli.StringFlag{
					Name:  "download_path, d",
					Usage: "Download destination file path. If empty string is speficied, the path will be ./screenshots.zip",
				},
				cli.StringFlag{
					Name:  "file_index_type, i",
					Usage: "'line_number' or 'auto_increment'. If empty string is specified, the type will be 'line_number'",
				},
				cli.StringFlag{
					Name:  "file_name_body_type, B",
					Usage: "'none' or 'screenshot_name'. If empty string is specified, the type will be 'none'",
				},
				cli.StringFlag{
					Name:  "download_type, D",
					Usage: "'all' or 'command_only' (i.e. screenshots only for 'Take screenshot' command). If empty string is specified, the type will be 'all'",
				},
				cli.BoolFlag{
					Name:  "mask_dynamically_changed_area, m",
					Usage: "Mask dynamically changed areas which can cause unexpected image difference between each test",
				},
				cli.IntFlag{
					Name:  "wait_limit, w",
					Usage: "Wait limit in seconds. The default value is 300",
				},
				cli.BoolFlag{
					Name:  "quiet, q",
					Usage: "Not output any logs during download. Disabled by default",
				},
			}...),
			Action: getScrenshotsAction,
		},
		{
			Name:  "wait-for-batch-run",
			Usage: "Wait until a batch run ends",
			Flags: append(commonFlags(), []cli.Flag{
				cli.IntFlag{
					Name:  "batch_run_number, b",
					Usage: "Batch run number",
				},
				cli.IntFlag{
					Name:  "wait_limit, w",
					Usage: "Wait limit in seconds. If 0 is specified, the value is test count x 10 minutes",
				},
			}...),
			Action: waitForBatchRunAction,
		},
	}
	app.Run(os.Args)
}

func getBatchRunAction(c *cli.Context) error {
	urlBase, apiToken, organization, project, httpHeadersMap, err := parseCommonFlags(c)
	if err != nil {
		return err
	}

	batchRunNumber := c.Int("batch_run_number")
	if batchRunNumber == 0 {
		return cli.NewExitError("--batch_run_number option is not specified or 0", 1)
	}
	batchRun, exitErr := common.GetBatchRun(urlBase, apiToken, organization, project, httpHeadersMap, batchRunNumber)
	if exitErr != nil {
		return exitErr
	}
	b, err := json.Marshal(batchRun)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func getBatchRunsAction(c *cli.Context) error {
	urlBase, apiToken, organization, project, httpHeadersMap, err := parseCommonFlags(c)
	if err != nil {
		return err
	}

	count := c.Int("count")
	maxBatchRunNumber := c.Int("max_batch_run_number")
	minBatchRunNumber := c.Int("min_batch_run_number")
	if maxBatchRunNumber != 0 && minBatchRunNumber != 0 && maxBatchRunNumber < minBatchRunNumber {
		return cli.NewExitError("--max_batch_run_number value is smaller than --min_batch_run_number value", 1)
	}
	batchRuns, exitErr := common.GetBatchRuns(urlBase, apiToken, organization, project, httpHeadersMap, count, maxBatchRunNumber, minBatchRunNumber)
	if exitErr != nil {
		return exitErr
	}
	b, err := json.Marshal(batchRuns)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func latestBatchRunNoAction(c *cli.Context) error {
	// handle command line arguments
	urlBase, apiToken, organization, project, httpHeadersMap, err := parseCommonFlags(c)
	if err != nil {
		return err
	}

	batchRunNo, exitErr := common.LatestBatchRunNo(urlBase, apiToken, organization, project, httpHeadersMap)
	if exitErr != nil {
		return exitErr
	}
	fmt.Printf("%d\n", batchRunNo)
	return nil
}

func uploadAppAction(c *cli.Context) error {
	// handle command line arguments
	urlBase, apiToken, organization, project, httpHeadersMap, err := parseCommonFlags(c)
	if err != nil {
		return err
	}
	appPath := c.String("app_path")
	if appPath == "" {
		return cli.NewExitError("--app_path option is required", 1)
	}

	fileNo, exitErr := common.UploadApp(urlBase, apiToken, organization, project, httpHeadersMap, appPath)
	if exitErr != nil {
		return exitErr
	}
	fmt.Printf("%d\n", fileNo)
	return nil
}

func deleteAppAction(c *cli.Context) error {
	// handle command line arguments
	urlBase, apiToken, organization, project, httpHeadersMap, err := parseCommonFlags(c)
	if err != nil {
		return err
	}
	appFileNumber := c.Int("app_file_number")
	if appFileNumber == 0 {
		return cli.NewExitError("--app_file_number option is not specified or 0", 1)
	}
	exitErr := common.DeleteApp(urlBase, apiToken, organization, project, httpHeadersMap, appFileNumber)
	if exitErr != nil {
		return exitErr
	}
	return nil
}

func getScrenshotsAction(c *cli.Context) error {
	// handle command line arguments
	urlBase, apiToken, organization, project, httpHeadersMap, err := parseCommonFlags(c)
	if err != nil {
		return err
	}
	batchRunNumber := c.Int("batch_run_number")
	if batchRunNumber == 0 {
		return cli.NewExitError("--batch_run_number option is not specified or 0", 1)
	}
	downloadPath := c.String("download_path")
	if downloadPath == "" {
		curDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		downloadPath = filepath.Join(curDir, "screenshots.zip")
	} else {
		stat, err := os.Stat(downloadPath)
		if err == nil {
			// downloadPath already exists
			mode := stat.Mode()
			if mode.IsDir() {
				return cli.NewExitError(fmt.Sprintf("'%s' should be not a directory but a file", downloadPath), 1)
			}
		}
	}
	downloadPath, err = filepath.Abs(downloadPath)
	if err != nil {
		panic(err)
	}
	fileIndexType := c.String("file_index_type")
	if fileIndexType == "" {
		fileIndexType = "line_number"
	}
	fileNameBodyType := c.String("file_name_body_type")
	if fileNameBodyType == "" {
		fileNameBodyType = "none"
	}
	downloadType := c.String("download_type")
	if downloadType == "" {
		downloadType = "all"
	}
	maskDynamicallyChangedArea := c.Bool("mask_dynamically_changed_area")
	waitLimit := c.Int("wait_limit")
	if waitLimit == 0 {
		waitLimit = -1
	}
	quiet := c.Bool("quiet")
	exitErr := common.GetScreenshots(urlBase, apiToken, organization, project, httpHeadersMap, batchRunNumber, downloadPath, fileIndexType, fileNameBodyType, downloadType, maskDynamicallyChangedArea, waitLimit, !quiet)
	if exitErr != nil {
		return exitErr
	}
	return nil
}

func batchRunAction(c *cli.Context) error {
	// handle command line arguments
	urlBase, apiToken, organization, project, httpHeadersMap, err := parseCommonFlags(c)
	if err != nil {
		return err
	}
	testSettingsNumber := c.Int("test_settings_number")
	setting := c.String("setting")
	if testSettingsNumber == 0 && setting == "" {
		return cli.NewExitError("Either of --test_settings_number or --setting option is required", 1)
	}
	noWait := c.Bool("no_wait")
	waitLimit := c.Int("wait_limit")

	_, existsErr, existsUnresolved, batchRunError := common.ExecuteBatchRun(urlBase, apiToken, organization,
		project, httpHeadersMap, testSettingsNumber, setting, !noWait, waitLimit, true)
	if batchRunError != nil {
		return batchRunError
	}
	if existsErr {
		return cli.NewExitError("", 1)
	}
	if existsUnresolved {
		return cli.NewExitError("", 2)
	}
	return nil
}

func waitForBatchRunAction(c *cli.Context) error {
	urlBase, apiToken, organization, project, httpHeadersMap, err := parseCommonFlags(c)
	if err != nil {
		return err
	}
	batchRunNumber := c.Int("batch_run_number")
	if batchRunNumber == 0 {
		return cli.NewExitError("--batch_run_number option is not specified or 0", 1)
	}
	waitLimit := c.Int("wait_limit")

	batchRunUnderProgress, batchRunError := common.GetBatchRun(urlBase, apiToken, organization, project, httpHeadersMap, batchRunNumber)
	if batchRunError != nil {
		return batchRunError
	}

	_, existsErr, existsUnresolved, batchRunError := common.WaitForBatchRunResult(urlBase, apiToken, organization,
		project, httpHeadersMap, batchRunUnderProgress, waitLimit, true)
	if batchRunError != nil {
		return batchRunError
	}
	if existsErr {
		return cli.NewExitError("", 1)
	}
	if existsUnresolved {
		return cli.NewExitError("", 2)
	}
	return nil
}

func commonFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "token, t",
			Usage:  "API token. You can get the value from https://app.magicpod.com/accounts/api-token/",
			EnvVar: "MAGICPOD_API_TOKEN",
		},
		cli.StringFlag{
			Name:   "organization, o",
			Usage:  "Organization name. (Not \"organization display name\", be careful!)",
			EnvVar: "MAGICPOD_ORGANIZATION",
		},
		cli.StringFlag{
			Name:   "project, p",
			Usage:  "Project name. (Not \"project display name\", be careful!)",
			EnvVar: "MAGICPOD_PROJECT",
		},
		cli.StringFlag{
			Name:  "http_headers, H",
			Usage: "Additional HTTP headers in JSON string format (e.g.: '{\"accept-language\":\"ja\"}')",
		},
	}
}

func parseCommonFlags(c *cli.Context) (string, string, string, string, map[string]string, error) {
	urlBase := c.GlobalString("url-base")
	apiToken := c.String("token")
	organization := c.String("organization")
	project := c.String("project")
	httpHeadersMap := make(map[string]string)
	var err error
	if urlBase == "" {
		err = cli.NewExitError("url-base argument cannot be empty", 1)
	} else if apiToken == "" {
		err = cli.NewExitError("--token option is required", 1)
	} else if organization == "" {
		err = cli.NewExitError("--organization option is required", 1)
	} else if project == "" {
		err = cli.NewExitError("--project option is required", 1)
	} else {
		err = nil
		httpHeadersStr := c.String("http_headers")
		if httpHeadersStr != "" {
			err = json.Unmarshal([]byte(httpHeadersStr), &httpHeadersMap)
			if err != nil {
				err = cli.NewExitError("http headers must be in JSON string format whose keys and values are string", 1)
			}
		}
	}
	return urlBase, apiToken, organization, project, httpHeadersMap, err
}
