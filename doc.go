/*
osquery is a distributed machine querying application that is built on raft.
Usage:
in a terminal window first run this

  $ osquery -http=":8081" -peers="http://127.0.0.1:8080,http://127.0.0.1:8081" -id 2

in a seperate termnial window run

  $ osquery -http=":8080" -peers="http://127.0.0.1:8080,http://127.0.0.1:8081" -id 1

you'll get something like this;

  READY!
  Enter Type (file_contains, file_exists, process_running):

follow the instructions.
*/
package main // import "sevki.org/osquery"
