/*
 *     Copyright 2024 The CNAI Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

const (
	// ArtifactTypeModelManifest specifies the artifact type for a model manifest.
	ArtifactTypeModelManifest = "application/vnd.cnai.model.manifest.v1+json"
)

const (
	// MediaTypeModelConfig specifies the media type for a model configuration.
	MediaTypeModelConfig = "application/vnd.cnai.model.config.v1+json"

	// MediaTypeModelWeightRaw is the media type used for an unarchived, uncompressed model weights.
	MediaTypeModelWeightRaw = "application/vnd.cnai.model.weight.v1.raw"

	// MediaTypeModelWeight is the media type used for model weights.
	MediaTypeModelWeight = "application/vnd.cnai.model.weight.v1.tar"

	// MediaTypeModelWeightGzip is the media type used for gzipped model weights.
	MediaTypeModelWeightGzip = "application/vnd.cnai.model.weight.v1.tar+gzip"

	// MediaTypeModelWeightZstd is the media type used for zstd compressed model weights.
	MediaTypeModelWeightZstd = "application/vnd.cnai.model.weight.v1.tar+zstd"

	// MediaTypeModelWeightConfigRaw is the media type used for an unarchived, uncompressed model weights, including files like `tokenizer.json`, `config.json`, etc.
	MediaTypeModelWeightConfigRaw = "application/vnd.cnai.model.weight.config.v1.raw"

	// MediaTypeModelConfig specifies the media type for configuration of the model weights, including files like `tokenizer.json`, `config.json`, etc.
	MediaTypeModelWeightConfig = "application/vnd.cnai.model.weight.config.v1.tar"

	// MediaTypeModelConfigGzip specifies the media type for gzipped configuration of the model weights, including files like `tokenizer.json`, `config.json`, etc.
	MediaTypeModelWeightConfigGzip = "application/vnd.cnai.model.weight.config.v1.tar+gzip"

	// MediaTypeModelConfigZstd specifies the media type for zstd compressed configuration of the model weights, including files like `tokenizer.json`, `config.json`, etc.
	MediaTypeModelWeightConfigZstd = "application/vnd.cnai.model.weight.config.v1.tar+zstd"

	// MediaTypeModelDocRaw is the media type used for an unarchived, uncompressed model documentation, including documentation files like `README.md`, `LICENSE`, etc.
	MediaTypeModelDocRaw = "application/vnd.cnai.model.doc.v1.raw"

	// MediaTypeModelDoc specifies the media type for model documentation, including documentation files like `README.md`, `LICENSE`, etc.
	MediaTypeModelDoc = "application/vnd.cnai.model.doc.v1.tar"

	// MediaTypeModelDocGzip specifies the media type for gzipped model documentation, including documentation files like `README.md`, `LICENSE`, etc.
	MediaTypeModelDocGzip = "application/vnd.cnai.model.doc.v1.tar+gzip"

	// MediaTypeModelDocZstd specifies the media type for zstd compressed model documentation, including documentation files like `README.md`, `LICENSE`, etc.
	MediaTypeModelDocZstd = "application/vnd.cnai.model.doc.v1.tar+zstd"

	// MediaTypeModelCodeRaw is the media type used for an unarchived, uncompressed model code, including code artifacts like scripts, code files etc.
	MediaTypeModelCodeRaw = "application/vnd.cnai.model.code.v1.raw"

	// MediaTypeModelCode specifies the media type for model code, including code artifacts like scripts, code files etc.
	MediaTypeModelCode = "application/vnd.cnai.model.code.v1.tar"

	// MediaTypeModelCodeGzip specifies the media type for gzipped model code, including code artifacts like scripts, code files etc.
	MediaTypeModelCodeGzip = "application/vnd.cnai.model.code.v1.tar+gzip"

	// MediaTypeModelCodeZstd specifies the media type for zstd compressed model code, including code artifacts like scripts, code files etc.
	MediaTypeModelCodeZstd = "application/vnd.cnai.model.code.v1.tar+zstd"

	// MediaTypeModelDatasetRaw is the media type used for an unarchived, uncompressed model datasets, including datasets that may be needed throughout the lifecycle of AI/ML models.
	MediaTypeModelDatasetRaw = "application/vnd.cnai.model.dataset.v1.raw"

	// MediaTypeModelDataset specifies the media type for model datasets, including datasets that may be needed throughout the lifecycle of AI/ML models.
	MediaTypeModelDataset = "application/vnd.cnai.model.dataset.v1.tar"

	// MediaTypeModelDatasetGzip specifies the media type for gzipped model datasets, including datasets that may be needed throughout the lifecycle of AI/ML models.
	MediaTypeModelDatasetGzip = "application/vnd.cnai.model.dataset.v1.tar+gzip"

	// MediaTypeModelDatasetZstd specifies the media type for zstd compressed model datasets, including datasets that may be needed throughout the lifecycle of AI/ML models.
	MediaTypeModelDatasetZstd = "application/vnd.cnai.model.dataset.v1.tar+zstd"
)
