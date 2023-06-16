A  quick start to developing the "website" project:

*Preconditions*
Install  go, make, node   
Install HUGO, note: you need to install the extension version


*Run the Harbor website locally*
Step 1: Clone project  and fork it to your github namespsce.
`git clone https://github.com/goharbor/website.git`
`cd website`
`git remote add $your_repo $your_forked_repo_URL`




Step 2: Load documentation content
The Markdown content for the Harbor docs is drawn from the docs folder and the release-X branches. To pull that content into your local website repo:

`make prepare`
This copies the docs directory and the release-X branches into this repo's content folder, separated by versions, where it can be processed by Hugo.

Step 3: Install npm dependencies
`npm i`
Step 4: Run Hugo in server mode
`make serve`
This starts up the local Hugo server on http://localhost:1313. As you make changes, the site refreshes automatically in your browser.





All the doc files are under the `docs` folder.


To view the changes you make,  you need to visit http://localhost:1313/docs/edge,  then navigate to the related section

