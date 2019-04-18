// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package job

// Define the register name constants of known jobs

const (
	// SampleJob is name of demo job
	SampleJob = "DEMO"

	// ImageScanJob is name of scan job it will be used as key to register to job service.
	ImageScanJob = "IMAGE_SCAN"
	// ImageScanAllJob is the name of "scanall" job in job service
	ImageScanAllJob = "IMAGE_SCAN_ALL"
	// ImageTransfer : the name of image transfer job in job service
	ImageTransfer = "IMAGE_TRANSFER"
	// ImageDelete : the name of image delete job in job service
	ImageDelete = "IMAGE_DELETE"
	// ImageReplicate : the name of image replicate job in job service
	ImageReplicate = "IMAGE_REPLICATE"
	// ImageGC the name of image garbage collection job in job service
	ImageGC = "IMAGE_GC"
)
