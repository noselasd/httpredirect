httpredirect
------------

Redirects all HTTP requests to another URL

    ./httpredirect -port 8080 -target http://myserver:8081/index


Build
-----
    cd $GOPATH
    go get github.com/noselasd/httpredirect
    go install github.com/noselasd/httpredirect
