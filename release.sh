#!/bin/bash
# SPDX-FileCopyrightText: 2021 Comcast Cable Communications Management, LLC
# SPDX-License-Identifier: Apache-2.0

version=${GITHUB_REF#refs/tags/}
meson=0

repo_slug=$1
changelog=$2
meson_wrap=$3
meson_provides=$4
artifact_dir=$5
verbose=0

if [ "$#" -eq 6 ]; then
    verbose=$6
fi

repo_tmp=(${repo_slug//\// })

org=${repo_tmp[0]}
repo_name=${repo_tmp[1]}

# Check the version to make sure we are running on a tag
raw_version=$version
if [[ "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+.*$ ]]; then
    if [ "v" == ${version:0:1} ]; then
        raw_version=${version:1}
    fi
else
    echo -e "\e[1;31mThis action only works on a version tag.  Examples: v1.0.2, 2.3.0, v99.2.3-release-12\e[0m"
    echo -e "Branch/tag executed on: \e[1;31m$version\e[0m"
    exit 1
fi

# Check for changelog file
if [[ -f $changelog ]]; then
    :
else
    echo -e "\e[1;31mThe changelog file ($changelog) is not present.\e[0m"
    exit 1
fi

# Input validate the meson_wrap
if [[ "$meson_wrap" =~ "true" ]]; then
    meson=1
elif [[ "$meson_wrap" =~ "false" ]]; then
    meson=0
else
    echo -e "\e[1;31mmeson-wrap must be true or false.\e[0m"
    exit 1
fi

release_slug=$(echo "$repo_name-$raw_version")

if [[ "none" == $meson_provides ]]; then
    meson_provides=$repo_name
fi

# This information should help if you are debugging something
if [ 0 -lt $verbose ]; then
    echo "       version: $version"
    echo "   raw_version: $raw_version"
    echo "     repo_slug: $repo_slug"
    echo "     repo_name: $repo_name"
    echo "  release_slug: $release_slug"

    if [ 1 == $meson ]; then
        echo "building meson wrap file"
        echo "meson_provides: $meson_provides"
    fi
fi

echo "Making the .tar.gz archive"
git archive --format=tar.gz -o $release_slug.tar.gz --prefix=$release_slug/ $version

echo "Making the .zip archive"
git archive --format=zip    -o $release_slug.zip --prefix=$release_slug/ $version

if [ 1 == $meson ]; then
    echo "Making the .wrap file for meson"

    tgz_sum=`sha256sum $release_slug.tar.gz`
    tgz_sum=${tgz_sum:0:64} # Keep only the first 64 bytes which are the checksum

    url="https://github.com/$repo_slug/releases/download/$version/$release_slug.tar.gz"

    echo "[wrap-file]"                                    > $repo_name.wrap
    echo "directory = $release_slug"                     >> $repo_name.wrap
    echo ""                                              >> $repo_name.wrap
    echo "source_filename = $release_slug.tar.gz"        >> $repo_name.wrap
    echo "source_url = $url"                             >> $repo_name.wrap
    echo "source_hash = $tgz_sum"                        >> $repo_name.wrap
    echo ""                                              >> $repo_name.wrap
    echo "[meson_provides]"                              >> $repo_name.wrap
    echo "lib$meson_provides = lib${meson_provides}_dep" >> $repo_name.wrap

    echo "Making the sha256sums.txt file"
    sha256sum $release_slug.tar.gz $release_slug.zip $repo_name.wrap > $release_slug-sha256sums.txt
else
    echo "Making the sha256sums.txt file"
    sha256sum $release_slug.tar.gz $release_slug.zip > $release_slug-sha256sums.txt
fi

echo "Copying files to the artifacts directory"
mkdir -p $artifact_dir
cp ${release_slug}* $artifact_dir/.

if [ -f $repo_name.wrap ]; then
    cp $repo_name.wrap $artifact_dir/.
fi

# Read through the changelog file and pull out the relavent release notes
output=0
notes=''
while read line; do
if [[ "$line" =~ ^##.*\[$version\].*$ ]]; then
    output=1
elif [[ "$line" =~ ^##.*\[v?[0-9]+\.[0-9]+\.[0-9]+.*\].*$ ]]; then
    output=0
fi

if [ 1 -eq $output ]; then
    notes=$(echo "${notes}\\n$line")
fi
done <$changelog

# Why YYYY-MM-DD?
# https://xkcd.com/1179/
today=`date +'%Y-%m-%d'`

# Defining the output variables
echo ::set-output name=release-name::$(echo ${version} ${today})
echo ::set-output name=release-body::${notes}
echo ::set-output name=artifact-dir::${artifact_dir}

exit 0
