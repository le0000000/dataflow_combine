# Dataflow vector combine
A simple example for combining vectors with Apache Beam Go pipeline. This repository is created to reproduce a bug we came across when doing combine operation on a PCollection.

## Issue
The issue only happens when running the pipeline with Google Dataflow, with some large data set. We are trying to combine a PCollection<pairedVec>, with
```
type pairedVec struct {
	Vec1 [1048576]uint64
	Vec2 [1048576]uint64
}
```
There are 10,000,000 items in the PCollection.

## How to run
```
bazel run -c opt pipeline:main_combine -- \
--input_file=gs://gcs_bucket/input/input_10m.txt \
--output_file=gs://gcs_bucket/output/output.txt \
--log_n=20 \
--runner=dataflow \
--project=gcp_id \
--region=us-east1 \
--temp_location=gs://gcs_bucket/tmp/ \
--staging_location=gs://gcs_bucket/binaries \
--worker_binary=/path_to_binary/main_combine

```
