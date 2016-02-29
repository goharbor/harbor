/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package controllers

// SearchController handles request to /search
type SearchController struct {
	BaseController
}

// Get renders page for displaying search result.
func (sc *SearchController) Get() {
	sc.Data["Username"] = sc.GetSession("username")
	sc.Data["QueryParam"] = sc.GetString("q")
	sc.ForwardTo("page_title_search", "search")
}
