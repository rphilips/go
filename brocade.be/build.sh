#!/bin/bash

export package=qtechng


export QTECHNG_BUILDTIME=$(date +%Y.%m.%d-%H:%M:%S)
export QTECHNG_GOVERSION=$(go version | cut -d " " -f 3)
export QTECHNG_BUILDHOST=$(hostname)

export STARTDIR=$(cd "$(dirname ../..)" >/dev/null; pwd -P) # $BROCADEGODIR
export QTECHNGBINDIR=`qtechng registry get qtechng-work-dir`/qtechng/binaries  # on workstation


platforms=( "linux/amd64"  "windows/amd64" "darwin/amd64" )
# platforms=( "linux/amd64" "darwin/amd64" )

for platform in "${platforms[@]}"
do
    echo $platform
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    basename=$package'-'$GOOS'-'$GOARCH

    cd $BROCADEGODIR/brocade.be/base
    env GOOS=$GOOS GOARCH=$GOARCH go install ./...

    cd $BROCADEGODIR/brocade.be/clipboard
    env GOOS=$GOOS GOARCH=$GOARCH go install

    cd $BROCADEGODIR/brocade.be/qtechng/cli

    env QTECHNG_BUILDTIME=$QTECHNG_BUILDTIME QTECHNG_GOVERSION=$QTECHNG_GOVERSION GOOS=$GOOS GOARCH=$GOARCH QTECHNG_BUILDHOST=$QTECHNG_BUILDHOST go build -o $BROCADEGODIR/brocade.be/qtechng/cli/$basename -ldflags "-X main.buildTime=$QTECHNG_BUILDTIME -X main.buildHost=$QTECHNG_BUILDHOST -X main.goVersion=$QTECHNG_GOVERSION" .

    echo $BROCADEGODIR/brocade.be/qtechng/cli/$basename
    mv $BROCADEGODIR/brocade.be/qtechng/cli/$basename $QTECHNGBINDIR

done



basename="qtechngw.exe"
cd $BROCADEGODIR/brocade.be/qtechng/cli
env QTECHNG_BUILDTIME=$QTECHNG_BUILDTIME QTECHNG_GOVERSION=$QTECHNG_GOVERSION GOOS=windows GOARCH=amd64 QTECHNG_BUILDHOST=$QTECHNG_BUILDHOST go build -o $BROCADEGODIR/brocade.be/qtechng/cli/$basename -ldflags "-H=windowsgui -X main.buildTime=$QTECHNG_BUILDTIME -X main.buildHost=$QTECHNG_BUILDHOST -X main.goVersion=$QTECHNG_GOVERSION" .


cp $QTECHNGBINDIR/qtechng-linux-amd64 /home/rphilips/bin/qtechng

echo "qtechng file ci --cwd=$QTECHNGBINDIR"






