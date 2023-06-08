# magicpod-api-client
Simple and useful wrapper for MagicPod Web API.

## Install

Download the latest magicpod-api-client executable from [here](https://github.com/Magic-Pod/magicpod-api-client/releases) and unzip it to any directory you prefer.

## Usage

- You can check the help by `./magicpod-api-client help`
- You can check the help of each command by `./magicpod-api-client help <command name>`

## Examples

In any cases below, you can check the test result status programatically by magicpod-api-client's return values.
- 0: Succeeded
- 1: Failed
- 2: Unresolved (Self-healing happened)

### Upload app, run batch test for the app, wait until the batch run is finished, and delete the app if the test passed.

```
export MAGICPOD_API_TOKEN=<API token displayed on https://app.magicpod.com/accounts/api-token/>
export MAGICPOD_ORGANIZATION=<organization>
export MAGICPOD_PROJECT=<project>
FILE_NO=$(./magicpod-api-client upload-app -a <path to app/ipa/apk>)
./magicpod-api-client batch-run -S <test_settings_number> -s "{\"app_file_number\":\"${FILE_NO}\"}"
if [ $? = 0 ]
then
  ./magicpod-api-client delete-app -a ${FILE_NO}
fi
```

### Run batch test for the app URL and return immediately

When you have already defined test settings on the project batch run page, the command is like below.
```
./magicpod-api-client batch-run -n -t <API token displayed on https://app.magicpod.com/accounts/api-token/> -o <organization> -p <project> -S <test_settings_number>
```

Or you can specify arbitrary settings.

```
./magicpod-api-client batch-run -n -t <API token displayed on https://app.magicpod.com/accounts/api-token/> -o <organization> -p <project> -s "{\"environment\":\"magic_pod\",\"os\":\"ios\",\"device_type\":\"simulator\",\"version\":\"13.1\",\"model\":\"iPhone 8\",\"app_type\":\"app_url\",\"app_url\":\"<URL to zipped app/ipa/apk>\"}"
```

### Run a multi-device pattern for the app URL, and wait until all batch runs are finished

```
./magicpod-api-client batch-run -t <API token displayed on https://app.magicpod.com/accounts/api-token/> -o <organization> -p <project> -S <test_settings_number>
```

Or

```
./magicpod-api-client batch-run -t <API token displayed on https://app.magicpod.com/accounts/api-token/> -o <organization> -p <project> -s "{\"test_settings\":[{\"environment\":\"magic_pod\",\"os\":\"ios\",\"device_type\":\"simulator\",\"version\":\"13.1\",\"model\":\"iPhone 8\",\"app_type\":\"app_url\",\"app_url\":\"<URL to zipped app/ipa/apk>\"},{\"environment\":\"magic_pod\",\"os\":\"ios\",\"device_type\":\"simulator\",\"version\":\"13.1\",\"model\":\"iPhone X\",\"app_type\":\"app_url\",\"app_url\":\"<URL to zipped app/ipa/apk>\"}]\,\"concurrency\": 1}"
```

You can execute the tests in parallel by settings `concurrency` a number greater than 1.

### Run 2 batch tests for different projects in parallel, and wait until all batch runs are finished

```
export MAGICPOD_API_TOKEN=<API token displayed on https://app.magicpod.com/accounts/api-token/>
export MAGICPOD_ORGANIZATION=<organization>

./magicpod-api-client batch-run -p <project_1> -S <test_settings_number_1> &
PID1=$!

./magicpod-api-client batch-run -p <project_2> -S <test_settings_number_2> &
PID2=$!

# return values of the wait commands will be magicpod-api-client's return values.
wait $PID1
wait $PID2
```

## Build from source

Run the following in the top directory of this repository.

```
go get -d .
go build
```

The following is the script to generate 64 bit executables and zip files for Mac, Linux, Windows.

```
# Mac64
GOOS=darwin GOARCH=amd64 go build -o out/mac64/magicpod-api-client
zip -jq out/mac64_magicpod-api-client.zip out/mac64/magicpod-api-client

# Linux64
GOOS=linux GOARCH=amd64 go build -o out/linux64/magicpod-api-client
zip -jq out/linux64_magicpod-api-client.zip out/linux64/magicpod-api-client

# Win64
GOOS=windows GOARCH=amd64 go build -o out/win64/magicpod-api-client.exe
zip -jq out/win64_magicpod-api-client.exe.zip out/win64/magicpod-api-client.exe
```

## Code-signing and Notarization of macOS binary

> **Note**: The similar process is not necessary for Windows and Linux binary.

To make a distributable binary for macOS, we need to code-sign and notarize the app. Otherwise, the app cannot be opened in the other machine due to macOS security feature to prevent malware.

### Prerequisites

- [ ] Store certificate for code-signing
- [ ] Create an app-specific password for the Apple ID

### Procedures

The basic procedures are as follows.

1. Build a macOS binary
2. Code-sign the binary
3. Zip the binary file before submission
4. Submit the zip file for notarization

> **Note**: You can check `<certificate name>` and `<Team ID>` by `security find-identity -v -p codesigning`.

The following is an example script.

```sh
# Build a macOS binary.
GOOS=darwin GOARCH=amd64 go build -o out/mac64/magicpod-api-client

# Code-sign the binary.
codesign -s <certificate name> -v --timestamp --options runtime out/mac64/magicpod-api-client

# For notarization, the binary file must be archived into a zip file.
zip -jq out/mac64_magicpod-api-client.zip out/mac64/magicpod-api-client

# Submit the zip file for notarization.
# You'll be asked to enter the app-specific password after this command.
xcrun notarytool submit out/mac64_magicpod-api-client.zip --apple-id <Apple ID> --team-id <Team ID> --wait
```

### References

- [Apple: Sign in to apps with your Apple ID using app-specific passwords](https://support.apple.com/en-us/HT204397)
- [Apple: Code Signing Tasks](https://developer.apple.com/library/archive/documentation/Security/Conceptual/CodeSigningGuide/Procedures/Procedures.html)
- [Apple: Upload your app to the notarization service](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution/customizing_the_notarization_workflow#3087734)
