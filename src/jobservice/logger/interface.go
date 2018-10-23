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
	Debug(v ...interface{})

	// For debuging with format
	Debugf(format string, v ...interface{})

	// For logging info
	Info(v ...interface{})

	// For logging info with format
	Infof(format string, v ...interface{})

	// For warning
	Warning(v ...interface{})

	// For warning with format
	Warningf(format string, v ...interface{})

	// For logging error
	Error(v ...interface{})

	// For logging error with format
	Errorf(format string, v ...interface{})

	// For fatal error
	Fatal(v ...interface{})

	// For fatal error with error
	Fatalf(format string, v ...interface{})
}
