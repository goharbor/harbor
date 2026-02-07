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

function generateSBOMs {
    # Generate SPDX SBOMs for all Harbor container images using Syft
    # Args: curTag baseTag sbomsDir images...
    local curTag=$1
    local baseTag=$2
    local sbomsDir=$3
    local images=${@:4}

    mkdir -p "$sbomsDir"

    for image in $images
    do
        local shortName=$(basename "$image")
        local sbomFile="$sbomsDir/${shortName}.spdx.json"

        echo "Generating SBOM for $image:$baseTag -> $sbomFile"
        retry 3 syft "$image:$baseTag" -o spdx-json="$sbomFile"
    done

    echo "Creating consolidated SBOM..."
    local consolidatedFile="$sbomsDir/harbor-sbom-${curTag}.spdx.json"
    mergeConsolidatedSBOM "$curTag" "$sbomsDir" "$consolidatedFile"
    echo "Consolidated SBOM written to $consolidatedFile"
}

function mergeConsolidatedSBOM {
    # Merge per-image SPDX JSONs into a single consolidated SPDX 2.3 document
    # Args: curTag sbomsDir outputFile
    local curTag=$1
    local sbomsDir=$2
    local outputFile=$3

    jq -n \
      --arg name "harbor-${curTag}" \
      --arg ns "https://github.com/goharbor/harbor/releases/${curTag}" \
      '{
        spdxVersion: "SPDX-2.3",
        dataLicense: "CC0-1.0",
        SPDXID: "SPDXRef-DOCUMENT",
        name: $name,
        documentNamespace: $ns,
        creationInfo: {
          created: (now | strftime("%Y-%m-%dT%H:%M:%SZ")),
          creators: ["Tool: syft", "Organization: goharbor"]
        },
        packages: [inputs.packages[]?],
        relationships: [inputs.relationships[]?],
        documentDescribes: [inputs.documentDescribes[]?]
      }' "$sbomsDir"/*.spdx.json > "$outputFile"
}

function signAndAttachSBOMs {
    # Sign container images with Cosign keyless signing and attach SBOMs
    # Must be called while logged into Docker Hub and GHCR
    # Args: curTag sbomsDir images 
    local curTag=$1
    local sbomsDir=$2
    local images=${@:3}

    for image in $images
    do
        local shortName=$(basename "$image")
        local sbomFile="$sbomsDir/${shortName}.spdx.json"
        local dockerHubRef="$image:$curTag"
        local ghcrRef="ghcr.io/$image:$curTag"

        echo "=== Processing $shortName ==="

        echo "Signing $dockerHubRef ..."
        retry 5 cosign sign --yes "$dockerHubRef"

        echo "Signing $ghcrRef ..."
        retry 5 cosign sign --yes "$ghcrRef"

        echo "Attaching SBOM to $dockerHubRef ..."
        retry 5 cosign attach sbom --type spdx --sbom "$sbomFile" "$dockerHubRef"

        echo "Attaching SBOM to $ghcrRef ..."
        retry 5 cosign attach sbom --type spdx --sbom "$sbomFile" "$ghcrRef"
    done
}