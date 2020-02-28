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

package api

//go:generate mockery -case snake -dir ../../api/artifact -name Controller -output ./artifact -outpkg artifact
//go:generate mockery -case snake -dir ../../api/scan -name Controller -output ./scan -outpkg scan
//go:generate mockery -case snake -dir ../../api/scan -name Checker -output ./scan -outpkg scan
//go:generate mockery -case snake -dir ../../api/scanner -name Controller -output ./scanner -outpkg scanner
