#!/bin/bash

# launches the webserver and opens the webpage

UIPORT=10000
WEBPORT=8080

URL="http://localhost:"
URL+=$WEBPORT


python -mwebbrowser $URL


# Lauch web server
./guiclient -UIPort=$UIPORT -port=$WEBPORT
