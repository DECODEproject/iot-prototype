# decode-prototype-da

Deliverable for D3.2

Architecture
------------

TODO

Global metadata service
Nodes holding their IOT data and entitlement data

Walkthrough
-----------

TODO


Notes
-----

- this is a prototype and not production software.
- there is no authentication and authentication.
- all data is held in memory - resetting the environment will reset all of the data

Building
--------

To build the software ensure you have installed the following software :

- Golang 1.7.3+
- Docker and Docker compose
- [Elm]( https://guide.elm-lang.org/install.html) 0.18+


Once you have a working installation download the code using the go get

```
go get gogs.dyne.org/DECODE/decode-prototype-da
```

The makefile contains helpers to build the environment plus some helpers for development.


```
make help
```

To build all of the docker containers locally for docker compose to use

```
make docker-build
```

To run the application components via docker compose

```
make docker-up
```

The prototype should then be available at the following urls 

'node' swagger api - http://localhost:8080/apidocs

'node' ui - http://localhost:8085/node.html

'metadata' swagger api - http://localhost:8081/apidocs

'metadata' ui - http://localhost:8085/search.html

'storage' swagger api - http://localhost:8083/apidocs

