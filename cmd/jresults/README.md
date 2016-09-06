# jresults

A tool for parsing jmeter result files and creating an HTML page with the
results in a more friendly format.

# Setup

You will need go-bindata if you are changing the html templates. Install with
the following command:

    $ go get -u github.com/jteeuwen/go-bindata/...

# Usage

    $ jresults path/to/results_file.jtl
