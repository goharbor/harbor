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

package lib

//go:generate mockery --case snake --dir ../../lib/orm --name Creator --output ./orm --outpkg orm
//go:generate mockery --case snake --dir ../../lib/cache --name Cache --output ./cache --outpkg cache
//go:generate mockery --case snake --dir ../../lib/cache --name Iterator --output ./cache --outpkg cache
//go:generate mockery --case snake --dir ../../lib/config --name Manager --output ./config --outpkg config
