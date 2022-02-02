#!/bin/bash

readarray -d . -t arr < .version
newVersion="${arr[0]}.${arr[1]}.$((${arr[2]}+1))" 
echo $newVersion > .version