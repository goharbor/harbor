#!/bin/bash
set -e

function getAssets {
    local bucket=$1
    local branch=$2
    local offlinePackage=$3
    local onlinePackage=$4
    local prerelease=$5
    local assetsPath=$6
    
    mkdir $assetsPath && pushd $assetsPath
    aws s3 cp s3://$bucket/$branch/$offlinePackage .
    md5sum $offlinePackage > md5sum
    # Pre-release does not handle online installer packages
    if [ "$prerelease" = "false" ]; then
        aws s3 cp s3://$bucket/$branch/$onlinePackage .
        md5sum $onlinePackage >> md5sum
    fi
    popd
}

function generateReleaseNotes {
    # Use .github/release.yml configuration to generate release notes for preTag to curTag
    local curTag=$1
    local preTag=$2
    local token=$3
    local releaseNotesPath=$4
    set +e
    # Calculate preTag if preTag is null
    # If curTag is v2.5.0-rc1 then preTag is v2.4.0
    # If curTag is v2.5.1 then preTag is v2.5.0
    if [ $preTag = "null" ]
    then
        IFS='.' read -r -a curTagArray <<< $curTag
        IFS='-' read -r -a patch <<< ${curTagArray[2]}
        local tagMajor=${curTagArray[0]}
        local tagMinor=${curTagArray[1]}
        local tagPatch=${patch[0]}
        if [ $tagPatch -gt 0 ]
        then
            preTag="$tagMajor.$tagMinor.$(expr $tagPatch - 1)"
        else
            preTag="$tagMajor.$(expr $tagMinor - 1).$tagPatch"
        fi
    fi
    set -e
    release=$(curl -X POST -H "Authorization: token $token" -H "Accept: application/vnd.github.v3+json" https://api.github.com/repos/goharbor/harbor/releases/generate-notes -d '{"tag_name":"'$curTag'","previous_tag_name":"'$preTag'"}' | jq '.body' | tr -d '"')
    echo -e $release > $releaseNotesPath
}

function publishImages {
    # Create curTag and push it to the goharbor namespace of dockerhub
    local curTag=$1
    local baseTag=$2
    local dockerHubUser=$3
    local dockerHubPassword=$4
    local images=${@:5}
    docker login -u $dockerHubUser -p $dockerHubPassword
    for image in $images
    do
        echo "push image: $image"
        docker tag $image:$baseTag $image:$curTag
        retry 5 docker push $image:$curTag
    done
    docker logout
}

function publishPackages {
    local curTag=$1
    local baseTag=$2
    local ghcrUser=$3
    local ghcrPassword=$4
    local images=${@:5}
    docker login ghcr.io -u $ghcrUser -p $ghcrPassword
    for image in $images
    do
        echo "push image: $image"
        docker tag $image:$baseTag "ghcr.io/"$image:$curTag
        retry 5 docker push "ghcr.io/"$image:$curTag
    done
    docker logout ghcr.io
}

function retry {
    local -r -i max="$1"; shift
    local -i n=1
    until "$@"
    do
        if ((n==max))
        then
            echo "fail with $n times try..."
            return 1
        else
            echo "failed, trying again in $n seconds..."
            sleep $((n++))
        fi
    done
}