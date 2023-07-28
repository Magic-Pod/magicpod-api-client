package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty"
	"github.com/mholt/archiver"
	"github.com/urfave/cli"
)

// BatchRun stands for a batch run executed on the server
type BatchRun struct {
	Url            string
	Status         string
	BatchRunNumber int
	TestCases      struct {
		Succeeded  int
		Failed     int
		Aborted    int
		Unresolved int
		Total      int
	}
}

// BatchRuns stands for a group of batch runs executed on the server
type BatchRuns struct {
	BatchRuns []BatchRun
}

// UploadFile stands for a file to be uploaded to the server
type UploadFile struct {
	FileNo int
}

func zipAppDir(dirPath string) string {
	zipPath := dirPath + ".zip"
	if err := os.RemoveAll(zipPath); err != nil {
		panic(err)
	}
	if err := archiver.Archive([]string{dirPath}, zipPath); err != nil {
		panic(err)
	}
	return zipPath
}

func createBaseRequest(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string) *resty.Request {
	client := resty.New()
	return client.
		SetHostURL(urlBase+"/api/v1.0").R().
		SetHeader("Authorization", "Token "+string(apiToken)).
		SetHeaders(httpHeadersMap).
		SetPathParams(map[string]string{
			"organization": organization,
			"project":      project,
		})
}

func handleError(resp *resty.Response) *cli.ExitError {
	if resp.StatusCode() != 200 {
		return cli.NewExitError(fmt.Sprintf("%s: %s", resp.Status(), resp.String()), 1)
	} else {
		return nil
	}
}

// UploadApp uploads app/ipa/apk file to the server
func UploadApp(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string, appPath string) (int, *cli.ExitError) {
	stat, err := os.Stat(appPath)
	if err != nil {
		return 0, cli.NewExitError(fmt.Sprintf("%s does not exist", appPath), 1)
	}
	var actualPath string
	if stat.Mode().IsDir() {
		if strings.HasSuffix(appPath, ".app") {
			actualPath = zipAppDir(appPath)
		} else {
			return 0, cli.NewExitError(fmt.Sprintf("%s is not file but direcoty.", appPath), 1)
		}
	} else {
		actualPath = appPath
	}
	res, err := createBaseRequest(urlBase, apiToken, organization, project, httpHeadersMap).
		SetFile("file", actualPath).
		SetResult(UploadFile{}).
		Post("/{organization}/{project}/upload-file/")
	if err != nil {
		panic(err)
	}
	if exitErr := handleError(res); exitErr != nil {
		return 0, exitErr
	}
	return res.Result().(*UploadFile).FileNo, nil
}

func mergeTestSettingsNumberToSetting(testSettingsMap map[string]interface{}, hasTestSettings bool, testSettingsNumber int) string {
	testSettingsMap["test_settings_number"] = testSettingsNumber

	if !hasTestSettings {
		// convert {\"model\":\"Nexus 5X\"} to {\"test_settings\":[{\"model\":\"Nexus 5X\"}]}
		// so that it can be treated with test_settings_number
		miscSettings := make(map[string]interface{})
		keysToDelete := []string{}
		for k, v := range testSettingsMap {
			if k != "test_settings_number" && k != "concurrency" && k != "test_settings_name" {
				miscSettings[k] = v
				keysToDelete = append(keysToDelete, k)
			}
		}
		for k := range keysToDelete {
			delete(testSettingsMap, keysToDelete[k])
		}
		if len(miscSettings) > 0 {
			settingsArray := [...]map[string]interface{}{miscSettings}
			testSettingsMap["test_settings"] = settingsArray
		}
	}

	settingBytes, _ := json.Marshal(testSettingsMap)
	return string(settingBytes)
}

// StartBatchRun starts a batch run or a cross batch run on the server
func StartBatchRun(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string, testSettingsNumber int, setting string) (*BatchRun, *cli.ExitError) {
	var testSettings interface{}
	isCrossBatchRunSetting := (testSettingsNumber != 0)
	if setting == "" {
		setting = "{\"test_settings_number\":" + strconv.Itoa(testSettingsNumber) + "}"
	} else {
		err := json.Unmarshal([]byte(setting), &testSettings)
		if err == nil {
			testSettingsMap, ok := testSettings.(map[string]interface{})
			if ok {
				_, hasTestSettings := testSettingsMap["test_settings"]
				testSettingsNumberInJSON, hasTestSettingsNumber := testSettingsMap["test_settings_number"]
				if testSettingsNumber != 0 {
					if hasTestSettingsNumber && testSettingsNumber != testSettingsNumberInJSON {
						return nil, cli.NewExitError("--test_settings_number and --setting have different number", 1)
					}
					setting = mergeTestSettingsNumberToSetting(testSettingsMap, hasTestSettings, testSettingsNumber)
				}
				isCrossBatchRunSetting = isCrossBatchRunSetting || hasTestSettings || hasTestSettingsNumber
			}
		}
	}
	if isCrossBatchRunSetting {
		res, err := createBaseRequest(urlBase, apiToken, organization, project, httpHeadersMap).
			SetHeader("Content-Type", "application/json").
			SetBody(setting).
			SetResult(BatchRun{}).
			Post("/{organization}/{project}/cross-batch-run/")
		if err != nil {
			panic(err)
		}
		if exitErr := handleError(res); exitErr != nil {
			return nil, exitErr
		}
		return res.Result().(*BatchRun), nil
	} else { // normal batch run
		res, err := createBaseRequest(urlBase, apiToken, organization, project, httpHeadersMap).
			SetHeader("Content-Type", "application/json").
			SetBody(setting).
			SetResult(BatchRun{}).
			Post("/{organization}/{project}/batch-run/")
		if err != nil {
			panic(err)
		}
		if exitErr := handleError(res); exitErr != nil {
			return nil, exitErr
		}
		return res.Result().(*BatchRun), nil
	}
}

// GetBatchRun retrieves status and number of test cases executed of a specified batch run
func GetBatchRun(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string, batchRunNumber int) (*BatchRun, *cli.ExitError) {
	res, err := createBaseRequest(urlBase, apiToken, organization, project, httpHeadersMap).
		SetPathParams(map[string]string{
			"batch_run_number": strconv.Itoa(batchRunNumber),
		}).
		SetResult(BatchRun{}).
		Get("/{organization}/{project}/batch-run/{batch_run_number}/")
	if err != nil {
		panic(err)
	}
	if exitErr := handleError(res); exitErr != nil {
		return nil, exitErr
	}
	return res.Result().(*BatchRun), nil
}

func LatestBatchRunNo(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string) (int, *cli.ExitError) {
	res, err := createBaseRequest(urlBase, apiToken, organization, project, httpHeadersMap).
		SetQueryParam("count", "1").
		SetResult(BatchRuns{}).
		Get("/{organization}/{project}/batch-runs/")
	if err != nil {
		panic(err)
	}
	if exitErr := handleError(res); exitErr != nil {
		return 0, exitErr
	}
	batchRuns := res.Result().(*BatchRuns).BatchRuns
	if len(batchRuns) == 0 {
		return 0, cli.NewExitError("no batch run exists in this project", 1)
	}
	return batchRuns[0].BatchRunNumber, nil
}

// DeleteApp deletes app/ipa/apk file on the server
func DeleteApp(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string, appFileNumber int) *cli.ExitError {
	res, err := createBaseRequest(urlBase, apiToken, organization, project, httpHeadersMap).
		SetBody(fmt.Sprintf("{\"app_file_number\":%d}", appFileNumber)).
		Delete("/{organization}/{project}/delete-file/")
	if err != nil {
		panic(err)
	}
	if exitErr := handleError(res); exitErr != nil {
		return exitErr
	}
	return nil
}

func PrepareScreenshots(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string, batchRunNumber int, fileIndexType string, fileNameBodyType string, downloadType string, maskDynamicallyChangedArea bool) int {
	var maskDynamicallyChangedAreaStr string
	if maskDynamicallyChangedArea {
		maskDynamicallyChangedAreaStr = "true"
	} else {
		maskDynamicallyChangedAreaStr = "false"
	}
	res, err := createBaseRequest(urlBase, apiToken, organization, project, httpHeadersMap).
		SetPathParams(map[string]string{
			"batch_run_number": strconv.Itoa(batchRunNumber),
		}).
		SetQueryParam("file_index_type", fileIndexType).
		SetQueryParam("file_name_body_type", fileNameBodyType).
		SetQueryParam("download_type", downloadType).
		SetQueryParam("mask_dynamically_changed_area", maskDynamicallyChangedAreaStr).
		SetResult(map[string]int{}).
		Post("/{organization}/{project}/batch-runs/{batch_run_number}/screenshots/")
	if err != nil {
		panic(err)
	}
	responseJson := res.Result().(*map[string]int)
	return (*responseJson)["batch_task_id"]
}

func GetBatchTaskStatus(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string, batchTaskId int) string {
	res, err := createBaseRequest(urlBase, apiToken, organization, project, httpHeadersMap).
		SetPathParams(map[string]string{
			"batch_task_id": strconv.Itoa(batchTaskId),
		}).
		SetResult(map[string]string{}).
		Get("/{organization}/{project}/batch-task/{batch_task_id}/")
	if err != nil {
		panic(err)
	}
	responseJson := res.Result().(*map[string]string)
	return (*responseJson)["status"]
}

func DownloadPreparedScreenshots(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string, batchTaskId int, downloadPath string) error {
	res, err := createBaseRequest(urlBase, apiToken, organization, project, httpHeadersMap).
		SetPathParams(map[string]string{
			"batch_task_id": strconv.Itoa(batchTaskId),
		}).
		SetResult(map[string]string{}).
		SetOutput(downloadPath).
		Get("/{organization}/{project}/screenshots/{batch_task_id}/")
	if err != nil {
		panic(err)
	}

	if res.StatusCode() != 200 {
		// response body is included not in res but in downloadPath file,
		responseText, err := ioutil.ReadFile(downloadPath)
		if err != nil {
			panic(err)
		}
		// remove downloadPath since it contains not zip contents but just error information
		err = os.Remove(downloadPath)
		if err != nil {
			panic(err)
		}
		return cli.NewExitError(fmt.Sprintf("%s: %s", res.Status(), responseText), 1)
	}
	return nil
}

func GetScreenshots(urlBase string, apiToken string, organization string, project string, httpHeadersMap map[string]string,
	batchRunNumber int, downloadPath string, fileIndexType string, fileNameBodyType string, downloadType string,
	maskDynamicallyChangedArea bool, waitLimit int, printResult bool) error {
	batchTaskId := PrepareScreenshots(urlBase, apiToken, organization, project, httpHeadersMap, batchRunNumber, fileIndexType, fileNameBodyType, downloadType, maskDynamicallyChangedArea)
	printMessage(printResult, "Preparing screenshots download.. \n")
	interval := 5
	passedSeconds := 0
	actualWaitLimit := waitLimit
	defaultTimeout := 300
	if waitLimit == -1 {
		actualWaitLimit = defaultTimeout
	}
	for {
		status := GetBatchTaskStatus(urlBase, apiToken, organization, project, httpHeadersMap, batchTaskId)
		if status == "succeeded" {
			printMessage(printResult, "\nDone.\n")
			break
		} else if status == "running" {
			printMessage(printResult, ".")
		} else {
			return cli.NewExitError("\nScreenshots download failed unexpectedly", 1)
		}
		if passedSeconds > 60 {
			interval = 10
		}
		if passedSeconds > 120 {
			interval = 30
		}
		if passedSeconds >= actualWaitLimit {
			errorMessage := fmt.Sprintf("\nReached timeout of %d seconds while waiting for screenshots download.", actualWaitLimit)
			if waitLimit == -1 {
				errorMessage += fmt.Sprintf("  Default timeout is %d seconds.  If it's not enough, please specify a longer value by --wait_limit or -w option.", defaultTimeout)
			}
			return cli.NewExitError(errorMessage, 1)
		}
		time.Sleep(time.Duration(interval) * time.Second)
		passedSeconds += interval
	}
	return DownloadPreparedScreenshots(urlBase, apiToken, organization, project, httpHeadersMap, batchTaskId, downloadPath)
}

func printMessage(printResult bool, format string, args ...interface{}) {
	if printResult {
		fmt.Printf(format, args...)
	}
}

// ExecuteBatchRun starts batch run(s) and wait for its completion with showing progress
func ExecuteBatchRun(urlBase string, apiToken string, organization string, project string,
	httpHeadersMap map[string]string, testSettingsNumber int, setting string,
	waitForResult bool, waitLimit int, printResult bool) (*BatchRun /*on which magicpod bitrise step depends */, bool, bool, *cli.ExitError) {
	// send batch run start request
	batchRun, exitErr := StartBatchRun(urlBase, apiToken, organization, project, httpHeadersMap, testSettingsNumber, setting)
	if exitErr != nil {
		return nil, false, false, exitErr
	}

	printMessage(printResult, "test result page:\n")
	printMessage(printResult, "%s\n", batchRun.Url)

	// finish before the test finish
	if !waitForResult {
		return batchRun, false, false, nil
	}

	return WaitForBatchRunResult(urlBase, apiToken, organization, project, httpHeadersMap, batchRun, waitLimit, printResult)
}

func WaitForBatchRunResult(urlBase string, apiToken string, organization string, project string,
	httpHeadersMap map[string]string, batchRun *BatchRun,
	waitLimit int, printResult bool) (*BatchRun /*on which magicpod bitrise step depends */, bool, bool, *cli.ExitError) {

	crossBatchRunTotalTestCount := batchRun.TestCases.Total
	const initRetryInterval = 10 // retry more frequently at first
	const retryInterval = 60
	var limitSeconds int
	if waitLimit == 0 {
		limitSeconds = crossBatchRunTotalTestCount * 10 * 60 // wait up to test count x 10 minutes by default
	} else {
		limitSeconds = waitLimit
	}
	passedSeconds := 0
	existsErr := false
	existsUnresolved := false
	printMessage(printResult, "\n#%d wait until %d tests to be finished.. \n", batchRun.BatchRunNumber, batchRun.TestCases.Total)
	prevFinished := 0
	for {
		batchRunUnderProgress, exitErr := GetBatchRun(urlBase, apiToken, organization, project, httpHeadersMap, batchRun.BatchRunNumber)
		if exitErr != nil {
			if printResult {
				fmt.Print(exitErr)
			}
			existsErr = true
			break // give up the wait here
		}
		finished := batchRunUnderProgress.TestCases.Succeeded + batchRunUnderProgress.TestCases.Failed + batchRunUnderProgress.TestCases.Aborted + batchRunUnderProgress.TestCases.Unresolved
		printMessage(printResult, ".") // show progress to prevent "long time no output" error on CircleCI etc
		// output progress
		if finished != prevFinished {
			notSuccessfulCount := ""
			if batchRunUnderProgress.TestCases.Failed > 0 {
				notSuccessfulCount = fmt.Sprintf("%d failed", batchRunUnderProgress.TestCases.Failed)
			}
			if batchRunUnderProgress.TestCases.Unresolved > 0 {
				if notSuccessfulCount != "" {
					notSuccessfulCount += ", "
				}
				notSuccessfulCount += fmt.Sprintf("%d unresolved", batchRunUnderProgress.TestCases.Unresolved)
			}
			if notSuccessfulCount != "" {
				notSuccessfulCount = fmt.Sprintf(" (%s)", notSuccessfulCount)
			}
			printMessage(printResult, "%d/%d finished%s\n", finished, batchRun.TestCases.Total, notSuccessfulCount)
			prevFinished = finished
		}
		if batchRunUnderProgress.Status != "running" {
			if batchRunUnderProgress.TestCases.Unresolved > 0 {
				existsUnresolved = true
			}
			if batchRunUnderProgress.Status == "succeeded" {
				printMessage(printResult, "batch run succeeded\n")
				break
			} else if batchRunUnderProgress.Status == "failed" {
				if batchRunUnderProgress.TestCases.Failed > 0 {
					unresolved := ""
					if existsUnresolved {
						unresolved = fmt.Sprintf(", %d unresolved", batchRunUnderProgress.TestCases.Unresolved)
					}
					printMessage(printResult, "batch run failed (%d failed%s)\n", batchRunUnderProgress.TestCases.Failed, unresolved)
				} else {
					printMessage(printResult, "batch run failed\n")
				}
				existsErr = true
				break
			} else if batchRunUnderProgress.Status == "unresolved" {
				printMessage(printResult, "batch run unresolved (%d unresolved)\n", batchRunUnderProgress.TestCases.Unresolved)
				break
			} else if batchRunUnderProgress.Status == "aborted" {
				printMessage(printResult, "batch run aborted\n")
				existsErr = true
				break
			} else {
				panic(batchRunUnderProgress.Status)
			}
		}
		if passedSeconds > limitSeconds {
			return batchRun, existsErr, existsUnresolved, cli.NewExitError("batch run never finished", 1)
		}
		if passedSeconds < 120 {
			time.Sleep(initRetryInterval * time.Second)
			passedSeconds += initRetryInterval
		} else {
			time.Sleep(retryInterval * time.Second)
			passedSeconds += retryInterval
		}
	}
	return batchRun, existsErr, existsUnresolved, nil
}
