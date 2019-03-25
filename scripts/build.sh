#!/usr/bin/env bash

fastbuild=off
buildinfo=off
target=texturepacker
while getopts fbt: opt
do
    case "$opt" in
      f)  fastbuild=on;;
      b)  buildinfo=on;;
      t)  target="$OPTARG";;
      \?)		# unknown flag
      	  echo >&2 \
	  "usage: $0 [-f] [-b] [-t target] [target ...]"
	  exit 1;;
    esac
done
shift `expr $OPTIND - 1`

if [[ "$fastbuild" != on ]]
then
    # format
    echo "==> Formatting..."
    goimports -w $(find .. -type f -name '*.go' -not -path "../vendor/*")

    # mod
    echo "==> Module tidy and vendor..."
    go mod tidy
    go mod vendor
    go mod download

    # lint
    echo "==> Linting..."
    gometalinter	--vendor \
                    --fast \
                    --enable-gc \
                    --tests \
                    --aggregate \
                    --disable=gotype \
                    ../
fi

# build
echo "==> Building $target"

if [[ "$buildinfo" != off ]]
then
    BUILD_PKG=github.com/master-g/texturepacker/internal/buildinfo
    COMMIT_HASH=$(git rev-parse --short HEAD)
    BUILD_DATE=$(date +%Y-%m-%dT%TZ%z)
    echo "==> Commit hash:$COMMIT_HASH Date:$BUILD_DATE"

    LD_FLAGS="-X ${BUILD_PKG}.CommitHash=${COMMIT_HASH} -X ${BUILD_PKG}.BuildDate=${BUILD_DATE}"

    # echo "${LD_FLAGS}"
    go build -ldflags "${LD_FLAGS}" -o ../bin/${target} ../cmd/${target}
else
    go build -o ../bin/${target} ../cmd/${target}
fi
