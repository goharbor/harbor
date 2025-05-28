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

package logger

// Interface for logger.
type Interface interface {
	// For debuging
	Debug(v ...any)

	// For debuging with format
	Debugf(format string, v ...any)

	// For logging info
	Info(v ...any)

	// For logging info with format
	Infof(format string, v ...any)

	// For warning
	Warning(v ...any)

	// For warning with format
	Warningf(format string, v ...any)

	// For logging error
	Error(v ...any)

	// For logging error with format
	Errorf(format string, v ...any)

	// For fatal error
	Fatal(v ...any)

	// For fatal error with error
	Fatalf(format string, v ...any)
}
