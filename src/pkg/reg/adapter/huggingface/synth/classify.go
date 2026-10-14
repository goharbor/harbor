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

package synth

import (
	"path"
	"strings"
)

// The classification rules below are copied from modctl pkg/modelfile/constants.go at commit
// 5868a8c5cad6a64179875474296acdecac8b4b8c (Copyright 2025 The ModelPack Authors, Apache-2.0).
// They are copied instead of imported because the media type of every layer is part of the
// manifest digest: an upstream pattern change must never shift our digests silently, only
// through a bump of Version. Importing modctl's builder would also pull in hundreds of packages.

// weightFileSizeThreshold is modctl's 128 * humanize.MByte (SI megabytes).
const weightFileSizeThreshold int64 = 128 * 1000 * 1000

type fileType int

const (
	fileTypeWeightConfig fileType = iota
	fileTypeWeight
	fileTypeCode
	fileTypeDoc
)

var (
	weightConfigPatterns = []string{
		"*.json",
		"*.jsonl",
		"*.json5",
		"*.jsonc",
		"*.yaml",
		"*.yml",
		"*.toml",
		"*.ini",
		"*.config",
		"*.cfg",
		"*.conf",
		"*.properties",
		"*.props",
		"*.prop",
		"*.xml",
		"*.xsd",
		"*.rng",
		"*.modelcard",
		"*.meta",
		"*tokenizer.model*",
		"*.tiktoken",
		"vocab.txt",
		"merges.txt",
		"added_tokens.txt",
		"spiece.model",
		"sentencepiece*.model",
		"sentencepiece*.vocab",
		"tiktoken.model",
		"chat_template.jinja",
		"config.json.*",
		"*.hparams",
		"*.params",
		"*.hyperparams",
		"*.wandb",
		"*.mlflow",
		"*.tensorboard",
	}

	weightPatterns = []string{
		"*.safetensors",
		"*.bin",
		"*.bin.*",
		"*.pt",
		"*.pth",
		"*.mar",
		"*.pte",
		"*.pt2",
		"*.ptl",
		"*.tflite",
		"*.h5",
		"*.hdf",
		"*.hdf5",
		"*.pb",
		"*.meta",
		"*.data-*",
		"*.index",
		"*.gguf",
		"*.gguf.*",
		"*.ggml",
		"*.ggmf",
		"*.ggjt",
		"*.q4_0",
		"*.q4_1",
		"*.q5_0",
		"*.q5_1",
		"*.q8_0",
		"*.f16",
		"*.f32",
		"*.ckpt",
		"*.checkpoint",
		"*.dist_ckpt",
		"tensor[0-9]*_[0-9]*",
		"*.tensor",
		"*.weights",
		"*.state",
		"*.embedding",
		"*.vocab",
		"*.ot",
		"*.engine",
		"*.trt",
		"*.onnx",
		"*.onnx_data*",
		"*.msgpack",
		"*.model",
		"*.pkl",
		"*.pickle",
		"*.keras",
		"*.joblib",
		"*.npy",
		"*.npz",
		"*.nc",
		"*.mlmodel",
		"*.coreml",
		"*.mil",
		"*.mleap",
		"*.surml",
		"*.llamafile",
		"*.llamafile.*",
		"*.caffemodel",
		"*.prototxt",
		"*.dlc",
		"*.circle",
		"*.nb",
		"*.arrow",
		"*.parquet",
		"*.ftz",
		"*.ark",
		"*.db",
	}

	codePatterns = []string{
		"*.py",
		"*.ipynb",
		"*.sh",
		"*.patch",
		"*.c",
		"*.h",
		"*.hxx",
		"*.cpp",
		"*.cc",
		"*.cxx",
		"*.c++",
		"*.hpp",
		"*.hh",
		"*.h++",
		"*.java",
		"*.js",
		"*.mjs",
		"*.cjs",
		"*.jsx",
		"*.ts",
		"*.tsx",
		"*.go",
		"*.rs",
		"*.swift",
		"*.rb",
		"*.php",
		"*.scala",
		"*.kt",
		"*.kts",
		"*.r",
		"*.R",
		"*.m",
		"*.mm",
		"*.f",
		"*.f90",
		"*.f95",
		"*.f03",
		"*.f08",
		"*.jl",
		"*.lua",
		"*.pl",
		"*.pm",
		"*.cs",
		"*.vb",
		"*.dart",
		"*.groovy",
		"*.elm",
		"*.erl",
		"*.hrl",
		"*.ex",
		"*.exs",
		"*.hs",
		"*.lhs",
		"*.clj",
		"*.cljs",
		"*.cljc",
		"*.cl",
		"*.lisp",
		"*.lsp",
		"*.scm",
		"*.ss",
		"*.rkt",
		"*.sql",
		"*.psql",
		"*.mysql",
		"*.sqlite",
		"*.zig",
		"*.cu",
		"*.cuh",
		"*.bash",
		"*.zsh",
		"*.fish",
		"*.csh",
		"*.tcsh",
		"*.ksh",
		"*.ps1",
		"*.psm1",
		"*.psd1",
		"*.bat",
		"*.cmd",
		"*.vbs",
		"*.wsf",
		"*.applescript",
		"*.scpt",
		"*.awk",
		"*.sed",
		"*.expect",
		"*.env",
		"*.env.*",
		".env*",
		"Makefile*",
		"*.dockerfile",
		"Dockerfile*",
		"*.mk",
		"*.cmake",
		"CMakeLists.txt",
		"*.gradle",
		"*.gradle.kts",
		"build.gradle*",
		"settings.gradle*",
		"*.sbt",
		"*.mill",
		"*.bazel",
		"*.bzl",
		"BUILD*",
		"WORKSPACE*",
		"*.buck",
		"BUCK*",
		"*.ninja",
		"*.gyp",
		"*.gypi",
		"*.waf",
		"wscript*",
		"package.json",
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		"requirements*.txt",
		"Pipfile*",
		"pyproject.toml",
		"setup.cfg",
		"tox.ini",
		"poetry.lock",
		"Cargo.toml",
		"Cargo.lock",
		"go.mod",
		"go.sum",
		"composer.json",
		"composer.lock",
		"Gemfile*",
		"*.gemspec",
		"mix.exs",
		"mix.lock",
		"rebar.config",
		"rebar.lock",
		"*.so",
		"*.dll",
		"*.dylib",
		"*.lib",
		"*.a",
	}

	docPatterns = []string{
		"*.txt",
		"*.md",
		"*.pdf",
		"LICENSE*",
		"README*",
		"SETUP*",
		"*requirements*",
		"*.log",
		"*.tfevents*",
		"*.doc",
		"*.docx",
		"*.docm",
		"*.dot",
		"*.dotx",
		"*.dotm",
		"*.rtf",
		"*.odt",
		"*.ott",
		"*.fodt",
		"*.pages",
		"*.wpd",
		"*.xls",
		"*.xlsx",
		"*.xlsm",
		"*.xlsb",
		"*.xlt",
		"*.xltx",
		"*.xltm",
		"*.ods",
		"*.ots",
		"*.fods",
		"*.numbers",
		"*.csv",
		"*.ppt",
		"*.pptx",
		"*.pptm",
		"*.pps",
		"*.ppsx",
		"*.ppsm",
		"*.pot",
		"*.potx",
		"*.potm",
		"*.odp",
		"*.otp",
		"*.fodp",
		"*.key",
		"*.epub",
		"*.mobi",
		"*.azw",
		"*.azw3",
		"*.fb2",
		"*.fb3",
		"*.lit",
		"*.pdb",
		"*.djvu",
		"*.djv",
		"*.html",
		"*.htm",
		"*.xhtml",
		"*.mhtml",
		"*.mht",
		"*.xml",
		"*.xsl",
		"*.xslt",
		"*.tex",
		"*.latex",
		"*.ltx",
		"*.bib",
		"*.rst",
		"*.asciidoc",
		"*.adoc",
		"*.textile",
		"*.wiki",
		"*.mediawiki",
		"*.org",
		"*.texi",
		"*.texinfo",
		"*.info",
		"*.man",
		"*.chm",
		"*.hlp",
		"*.xps",
		"*.jpg",
		"*.jpeg",
		"*.png",
		"*.gif",
		"*.bmp",
		"*.tiff",
		"*.ico",
		"*.webp",
		"*.heic",
		"*.heif",
		"*.hevc",
		"*.svg",
		"*.mp4",
		"*.mov",
		"*.avi",
		"*.mkv",
		"*.webm",
		"*.m4v",
		"*.flv",
		"*.wmv",
		"*.mpg",
		"*.mpeg",
	}
)

// classify returns the file type of a repository path. Only the basename is matched, case-insensitively,
// in the order weight config, weight, code, doc. Unmatched files are weights above the size threshold,
// code otherwise.
func classify(filePath string, size int64) fileType {
	switch base := strings.ToLower(path.Base(filePath)); {
	case matchAny(base, weightConfigPatterns):
		return fileTypeWeightConfig
	case matchAny(base, weightPatterns):
		return fileTypeWeight
	case matchAny(base, codePatterns):
		return fileTypeCode
	case matchAny(base, docPatterns):
		return fileTypeDoc
	case size > weightFileSizeThreshold:
		return fileTypeWeight
	default:
		return fileTypeCode
	}
}

func matchAny(lowerBase string, patterns []string) bool {
	for _, p := range patterns {
		if ok, err := path.Match(strings.ToLower(p), lowerBase); err == nil && ok {
			return true
		}
	}
	return false
}

func (t fileType) mediaType() string {
	switch t {
	case fileTypeWeightConfig:
		return MediaTypeWeightConfigRaw
	case fileTypeWeight:
		return MediaTypeWeightRaw
	case fileTypeDoc:
		return MediaTypeDocRaw
	default:
		return MediaTypeCodeRaw
	}
}
