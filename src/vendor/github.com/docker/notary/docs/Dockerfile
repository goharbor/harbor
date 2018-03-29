FROM docs/base:oss
MAINTAINER Mary Anthony <mary@docker.com> (@moxiegirl)

ENV PROJECT=notary

# To get the git info for this repo
COPY . /src
RUN rm -r /docs/content/$PROJECT/
COPY . /docs/content/$PROJECT/
