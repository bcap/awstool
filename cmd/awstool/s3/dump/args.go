package dump

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
)

var region string
var bucket string
var downloadVersions bool
var maxVersions int = 100
var oldestVersionStr string
var oldestVersion time.Time
var outputDir string

var durationPattern = regexp.MustCompile(`` +
	`^` +
	`(?:(\d+)d)?` + // optional days
	`(?:(\d+)h)?` + // optional hours
	`(?:(\d+)m)?` + // optional minutes
	`(?:(\d+)s)?` + // optional seconds
	`$`,
)

func Command(awsCfg **aws.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:           "dump",
		Short:         "downloads all files from a given s3 bucket",
		SilenceErrors: true,
		PreRunE:       parseArgs,
		RunE:          run,
	}

	cmd.PersistentFlags().StringVarP(
		&region, "region", "r", region,
		"Which region to operate on",
	)
	cmd.PersistentFlags().StringVarP(
		&bucket, "bucket", "b", bucket,
		"Which bucket to download files from",
	)
	cmd.PersistentFlags().BoolVar(
		&downloadVersions, "with-versions", downloadVersions,
		"Also download previous versions of the bucket items. Check also --max-versions and --oldest-version",
	)
	cmd.PersistentFlags().IntVar(
		&maxVersions, "max-versions", maxVersions,
		"When --with-versions is enabled, how many previous versions to download at most for each bucket item. "+
			"Set to 0 to unlimited versions",
	)
	cmd.PersistentFlags().StringVar(
		&oldestVersionStr, "oldest-version", oldestVersionStr,
		"When --with-versions is enabled, only downloads versions that came after this time. "+
			"Accepts either a timestamp (eg \"2006-01-02 15:04:05\") or a duration (eg \"15d3h10m\")",
	)
	cmd.PersistentFlags().StringVarP(
		&outputDir, "output-dir", "o", outputDir,
		"Local directory to store files. If not specified, files will be downloaded to ./<bucket>",
	)

	return &cmd
}

func parseArgs(cmd *cobra.Command, args []string) error {
	if region == "" {
		return errors.New("region not specified")
	}
	if bucket == "" {
		return errors.New("bucket not specified")
	}
	if maxVersions < 0 {
		return fmt.Errorf("max versions cannot be lower than zero: %d", maxVersions)
	}
	if oldestVersionStr != "" {
		if err := parseOldest(); err != nil {
			return err
		}
	}
	if outputDir == "" {
		outputDir = bucket
	}
	return nil
}

func parseOldest() error {
	validTime := func() error {
		if time.Since(oldestVersion) < 0 {
			return fmt.Errorf("oldest version is after current time: %v", oldestVersion)
		}
		return nil
	}

	var err error

	// try parsing as date + time
	oldestVersion, err = time.Parse("2006-01-02 15:04:05", oldestVersionStr)
	if err == nil {
		return validTime()
	}

	// try parsing as date only
	oldestVersion, err = time.Parse("2006-01-02", oldestVersionStr)
	if err == nil {
		return validTime()
	}

	// try parsing as duration
	submatches := durationPattern.FindStringSubmatch(oldestVersionStr)
	if submatches == nil {
		return fmt.Errorf("unknown oldest version: %s", oldestVersionStr)
	}
	days, _ := strconv.Atoi(submatches[1])
	hours, _ := strconv.Atoi(submatches[2])
	minutes, _ := strconv.Atoi(submatches[3])
	seconds, _ := strconv.Atoi(submatches[4])
	delta := time.Duration(days*24)*time.Hour +
		time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second
	if delta == 0 {
		return fmt.Errorf("0 duration specified: %s", oldestVersionStr)
	}
	oldestVersion = time.Now().Add(-delta)

	return nil
}
