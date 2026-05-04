// Package main provides the `lambdas-packer` CLI.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/felipefuhr/ffreis-lambdas-packer/internal/packer"
)

type options struct {
	bucket      string
	prefix      string
	artifactDir string
	region      string
	dryRun      bool
	noDelete    bool
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	opts, err := parseArgs(args)
	if err != nil {
		writeLine(os.Stderr, err.Error())
		return 2
	}

	ctx := context.Background()

	awsCfg, err := loadAWSConfig(ctx, opts.region)
	if err != nil {
		writeErrorLine(os.Stderr, "failed to load AWS config: ", err)
		return 1
	}

	prefix, err := packer.NormalizePrefix(opts.prefix)
	if err != nil {
		writeLine(os.Stderr, err.Error())
		return 2
	}

	local, err := packer.DiscoverLocalArtifacts(opts.artifactDir, prefix)
	if err != nil {
		writeErrorLine(os.Stderr, "artifact discovery failed: ", err)
		return 1
	}

	s3Client := s3.NewFromConfig(awsCfg)
	remote, err := packer.ListRemoteZips(ctx, s3Client, opts.bucket, prefix)
	if err != nil {
		writeErrorLine(os.Stderr, "failed listing s3://"+opts.bucket+"/"+prefix+": ", err)
		return 1
	}

	plan := packer.BuildPlan(local, remote, opts.noDelete)
	printPlan(plan, opts.bucket, prefix, opts.dryRun)

	if opts.dryRun {
		return 0
	}

	for _, a := range plan.Uploads {
		if err := packer.PutArtifact(ctx, s3Client, opts.bucket, a); err != nil {
			writeErrorLine(os.Stderr, "upload failed for "+a.Key+": ", err)
			return 1
		}
	}
	if err := packer.DeleteKeys(ctx, s3Client, opts.bucket, plan.Deletes); err != nil {
		writeErrorLine(os.Stderr, "delete failed: ", err)
		return 1
	}

	writeLine(os.Stdout, "done: uploaded="+strconv.Itoa(len(plan.Uploads))+" deleted="+strconv.Itoa(len(plan.Deletes)))
	return 0
}

func parseArgs(args []string) (options, error) {
	var opts options
	fs := flag.NewFlagSet(cliName, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&opts.bucket, flagBucket, "", "S3 bucket name (required)")
	fs.StringVar(&opts.prefix, flagPrefix, "", "S3 key prefix (required, non-empty; e.g. lambdas/dev/)")
	fs.StringVar(&opts.artifactDir, flagArtifactDir, "lambdas/target/lambda", "Artifact dir containing */bootstrap.zip or */bootstrap")
	fs.StringVar(&opts.region, flagRegion, "", "AWS region override (optional)")
	fs.BoolVar(&opts.dryRun, flagDryRun, false, "Print planned actions without changing S3")
	fs.BoolVar(&opts.noDelete, flagNoDelete, false, "Upload/update only (do not delete remote extras)")

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}

	if opts.bucket == "" {
		return options{}, fmt.Errorf("--bucket is required")
	}
	if opts.prefix == "" {
		return options{}, fmt.Errorf("--prefix is required and must be non-empty")
	}
	return opts, nil
}

func loadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
	if region != "" {
		return config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}
	return config.LoadDefaultConfig(ctx)
}

func printPlan(plan packer.Plan, bucket, prefix string, dryRun bool) {
	mode := modeApply
	if dryRun {
		mode = modeDryRun
	}
	writeLine(os.Stdout, cliName+" ("+mode+")")
	writeLine(os.Stdout, labelBucket+": "+bucket)
	writeLine(os.Stdout, labelPrefix+": "+prefix)
	writeLine(os.Stdout, "uploads: "+strconv.Itoa(len(plan.Uploads)))
	writeLine(os.Stdout, "deletes: "+strconv.Itoa(len(plan.Deletes)))
}

func writeLine(w io.Writer, line string) {
	_, _ = io.WriteString(w, line+"\n") //nolint:errcheck // writing to stdout/stderr is best-effort
}

func writeErrorLine(w io.Writer, prefix string, err error) {
	writeLine(w, prefix+err.Error())
}
