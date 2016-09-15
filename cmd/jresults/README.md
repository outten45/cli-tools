# jresults

A tool for parsing jmeter result files and creating an HTML page with the
results in a more friendly format.

# Setup

You will need go-bindata if you are changing the html templates. Install with
the following command:

    $ go get -u github.com/jteeuwen/go-bindata/...

If you make changes to the files in the assets directory, you will need to run
the following command:

    $ go generate

# Building

Used the build script to build a binary the includes the build stamp and
githash.

    $ ./build.sh

# Usage

    $ jresults path/to/results_file.jtl
